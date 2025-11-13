<template>
  <div class="reader-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>读者管理</span>
        </div>
      </template>

      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="请输入读者编号、姓名、身份证号进行搜索"
          clearable
          @input="handleSearch"
          style="width: 400px; margin-right: 10px;"
        >
          <template #prefix>
            <el-icon><SearchIcon /></el-icon>
          </template>
        </el-input>
      </div>

      <!-- 读者表格 -->
      <el-table
        :data="readerList"
        style="width: 100%; margin-top: 20px;"
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
        inactive: 'info',
        rejected: 'danger'
      }
      return map[status] || 'info'
    }

    const getStatusText = (status) => {
      const map = {
        pending: '待审核',
        active: '正常',
        inactive: '停用',
        rejected: '已拒绝'
      }
      return map[status] || status
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
      handleSubmit,
      handleDialogClose
    }
  }
}
</script>

<style scoped>
.reader-list {
  padding: 20px;
}

.search-bar {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

