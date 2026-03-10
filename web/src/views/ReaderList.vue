<template>
  <div class="reader-list">
    <el-card class="reader-manage-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div class="title-wrap">
            <span class="title-main">读者管理</span>
            <span class="title-sub">借阅权限、审核与黑名单状态总览</span>
          </div>
        </div>
      </template>

      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          class="reader-search-input"
          placeholder="请输入读者编号、姓名、身份证号进行搜索"
          clearable
          @input="handleSearch"
        >
          <template #prefix>
            <el-icon><SearchIcon /></el-icon>
          </template>
        </el-input>
      </div>

      <!-- 读者表格 -->
      <el-table
        class="reader-table"
        :data="readerList"
        v-loading="loading"
        border
      >
        <el-table-column label="ID" width="80">
          <template #default="scope">
            {{ scope.row.id || scope.row.ID || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="reader_no" label="读者编号" width="120" />
        <el-table-column prop="user.real_name" label="姓名" width="120" />
        <el-table-column prop="user.username" label="用户名" width="120" />
        <el-table-column prop="id_card" label="身份证号" width="180" />
        <el-table-column prop="address" label="地址" min-width="150" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag
              class="status-tag"
              :type="getStatusType(scope.row.status)"
            >
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="max_borrow" label="最大借阅数" width="120" />
        <el-table-column prop="borrow_days" label="借阅天数" width="120" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button
              v-if="scope.row.status === 'pending'"
              type="success"
              size="small"
              @click="handleApprove(scope.row)"
            >
              审核通过
            </el-button>
            <el-button
              v-if="scope.row.status === 'pending'"
              type="danger"
              size="small"
              @click="handleReject(scope.row)"
            >
              拒绝
            </el-button>
            <el-button
              type="primary"
              size="small"
              @click="handleEdit(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="scope.row.status === 'inactive' || scope.row.is_blacklisted"
              type="warning"
              size="small"
              @click="handleUnban(scope.row)"
            >
              解禁
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="编辑读者信息"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        :model="form"
        ref="formRef"
        label-width="100px"
      >
        <el-form-item label="最大借阅数">
          <el-input-number
            v-model="form.max_borrow"
            :min="1"
            :max="20"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="借阅天数">
          <el-input-number
            v-model="form.borrow_days"
            :min="7"
            :max="90"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="地址">
          <el-input
            v-model="form.address"
            type="textarea"
            :rows="3"
            placeholder="请输入地址"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import axios from 'axios'

export default {
  name: 'ReaderList',
  components: {
    SearchIcon: Search
  },
  setup() {
    const loading = ref(false)
    const readerList = ref([])
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(10)
    const searchKeyword = ref('')
    const dialogVisible = ref(false)
    const formRef = ref(null)

    const form = reactive({
      id: null,
      max_borrow: 5,
      borrow_days: 30,
      address: '',
      remark: ''
    })

    const getStatusType = (status) => {
      const map = {
        pending: 'warning',
        active: 'success',
        inactive: 'danger',
        rejected: 'danger'
      }
      return map[status] || 'info'
    }

    const getStatusText = (status) => {
      const map = {
        pending: '待审核',
        active: '正常',
        inactive: '已拉黑',
        rejected: '已拒绝'
      }
      return map[status] || status
    }

    const handleUnban = (row) => {
      ElMessageBox.confirm('确定要解除该用户的黑名单并恢复借阅权限吗？', '解禁确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const response = await axios.post(`/blacklist/removeByUser/${row.user_id}`, {
            remark: '管理员在读者管理页手动解禁'
          })
          if (response.code === 200) {
            ElMessage.success('解禁成功')
            fetchReaderList()
          } else {
            ElMessage.error(response.msg || '解禁失败')
          }
        } catch (error) {
          console.error('解禁失败:', error)
          ElMessage.error('解禁失败')
        }
      }).catch(() => {})
    }

    const fetchReaderList = async () => {
      loading.value = true
      try {
        const params = {
          page: currentPage.value,
          pageSize: pageSize.value
        }
        if (searchKeyword.value) {
          params.keyword = searchKeyword.value
        }

        const response = await axios.get('/reader/getReaderList', { params })
        if (response.code === 200) {
          readerList.value = response.data.list || []
          total.value = response.data.total || 0
        } else {
          ElMessage.error(response.msg || '获取数据失败')
        }
      } catch (error) {
        console.error('获取读者列表失败:', error)
        ElMessage.error('获取读者列表失败')
      } finally {
        loading.value = false
      }
    }

    const handleSearch = () => {
      currentPage.value = 1
      fetchReaderList()
    }

    const handleSizeChange = (val) => {
      pageSize.value = val
      currentPage.value = 1
      fetchReaderList()
    }

    const handlePageChange = (val) => {
      currentPage.value = val
      fetchReaderList()
    }

    const handleApprove = (row) => {
      ElMessageBox.confirm('确定要审核通过该读者吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const response = await axios.put('/reader/updateReaderStatus', {
            id: row.id || row.ID,
            status: 'active',
            remark: '审核通过'
          })
          if (response.code === 200) {
            ElMessage.success('审核通过')
            fetchReaderList()
          } else {
            ElMessage.error(response.msg || '操作失败')
          }
        } catch (error) {
          console.error('操作失败:', error)
          ElMessage.error('操作失败')
        }
      }).catch(() => {})
    }

    const handleReject = (row) => {
      ElMessageBox.confirm('确定要拒绝该读者吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const response = await axios.put('/reader/updateReaderStatus', {
            id: row.id || row.ID,
            status: 'rejected',
            remark: '审核拒绝'
          })
          if (response.code === 200) {
            ElMessage.success('已拒绝')
            fetchReaderList()
          } else {
            ElMessage.error(response.msg || '操作失败')
          }
        } catch (error) {
          console.error('操作失败:', error)
          ElMessage.error('操作失败')
        }
      }).catch(() => {})
    }

    const handleEdit = (row) => {
      Object.assign(form, {
        id: row.id || row.ID,
        max_borrow: row.max_borrow,
        borrow_days: row.borrow_days,
        address: row.address || '',
        remark: row.remark || ''
      })
      dialogVisible.value = true
    }

    const handleSubmit = async () => {
      try {
        const response = await axios.put('/reader/updateReader', form)
        if (response.code === 200) {
          ElMessage.success('更新成功')
          dialogVisible.value = false
          fetchReaderList()
        } else {
          ElMessage.error(response.msg || '更新失败')
        }
      } catch (error) {
        console.error('更新失败:', error)
        ElMessage.error('更新失败')
      }
    }

    const handleDialogClose = () => {
      Object.assign(form, {
        id: null,
        max_borrow: 5,
        borrow_days: 30,
        address: '',
        remark: ''
      })
    }

    onMounted(() => {
      fetchReaderList()
    })

    return {
      loading,
      readerList,
      total,
      currentPage,
      pageSize,
      searchKeyword,
      dialogVisible,
      form,
      formRef,
      getStatusType,
      getStatusText,
      handleSearch,
      handleSizeChange,
      handlePageChange,
      handleApprove,
      handleReject,
      handleEdit,
      handleUnban,
      handleSubmit,
      handleDialogClose
    }
  }
}
</script>

<style scoped>
.reader-list {
  padding: 18px;
}

.reader-manage-card {
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.54);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  box-shadow:
    0 12px 30px rgba(15, 23, 42, 0.12),
    0 24px 60px rgba(15, 23, 42, 0.08);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-wrap {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.title-main {
  font-size: 18px;
  font-weight: 800;
  color: #0f172a;
}

.title-sub {
  font-size: 12px;
  color: #64748b;
}

.search-bar {
  margin-bottom: 18px;
}

.reader-search-input {
  width: min(460px, 100%);
}

.reader-table {
  margin-top: 16px;
}

.reader-table :deep(.el-table__header-wrapper th),
.reader-table :deep(.el-table__body-wrapper td) {
  background: rgba(255, 255, 255, 0.46);
}

.reader-table :deep(.el-table__row:hover > td) {
  background: rgba(59, 130, 246, 0.08) !important;
}

.status-tag {
  border-radius: 999px;
  font-weight: 600;
  padding-inline: 10px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

:global(body.app-theme-dark) .reader-manage-card,
:global(.layout-container.theme-dark) .reader-manage-card {
  background: rgba(15, 23, 42, 0.52);
  border-color: rgba(148, 163, 184, 0.24);
  box-shadow: 0 18px 40px rgba(2, 6, 23, 0.48);
  backdrop-filter: none;
  -webkit-backdrop-filter: none;
}

:global(body.app-theme-dark) .title-main,
:global(.layout-container.theme-dark) .title-main {
  color: #f8fafc;
}

:global(body.app-theme-dark) .title-sub,
:global(.layout-container.theme-dark) .title-sub {
  color: #cbd5e1;
}

:global(body.app-theme-dark) .reader-table .el-table__header-wrapper th,
:global(body.app-theme-dark) .reader-table .el-table__body-wrapper td,
:global(.layout-container.theme-dark) .reader-table .el-table__header-wrapper th,
:global(.layout-container.theme-dark) .reader-table .el-table__body-wrapper td {
  background: rgba(15, 23, 42, 0.52);
  color: #e2e8f0;
}
</style>

