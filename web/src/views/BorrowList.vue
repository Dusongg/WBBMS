<template>
  <div class="borrow-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>借还管理</span>
          <div class="header-actions">
            <el-radio-group v-model="statusFilter" @change="handleFilterChange" style="margin-right: 10px;">
              <el-radio-button label="">全部</el-radio-button>
              <el-radio-button label="pending">待审批</el-radio-button>
              <el-radio-button label="borrowed">已借出</el-radio-button>
              <el-radio-button label="overdue">逾期</el-radio-button>
              <el-radio-button label="returned">已归还</el-radio-button>
            </el-radio-group>
            <el-button type="primary" @click="handleBorrow">借书</el-button>
          </div>
        </div>
      </template>

      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="请输入书名或ISBN进行搜索"
          clearable
          @input="handleSearch"
          style="width: 400px; margin-right: 10px;"
        >
          <template #prefix>
            <el-icon><SearchIcon /></el-icon>
          </template>
        </el-input>
      </div>

      <!-- 借阅记录表格 -->
      <el-table
        :data="borrowList"
        style="width: 100%; margin-top: 20px;"
        v-loading="loading"
        border
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="reader.user.real_name" label="读者" width="120" />
        <el-table-column prop="book.title" label="书名" min-width="150" />
        <el-table-column prop="book.isbn" label="ISBN" width="150" />
        <el-table-column prop="borrow_date" label="借阅日期" width="120">
          <template #default="scope">
            {{ formatDate(scope.row.borrow_date) }}
          </template>
        </el-table-column>
        <el-table-column prop="due_date" label="应还日期" width="120">
          <template #default="scope">
            {{ formatDate(scope.row.due_date) }}
          </template>
        </el-table-column>
        <el-table-column prop="return_date" label="归还日期" width="120">
          <template #default="scope">
            {{ scope.row.return_date ? formatDate(scope.row.return_date) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="fine_amount" label="罚款金额" width="100">
          <template #default="scope">
            ¥{{ scope.row.fine_amount ? scope.row.fine_amount.toFixed(2) : '0.00' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <!-- 待审批状态：显示批准/拒绝按钮 -->
            <div v-if="scope.row.status === 'pending'" style="display: flex; gap: 8px; justify-content: center;">
              <el-button
                type="success"
                size="small"
                @click="handleApprove(scope.row, true)"
              >
                批准
              </el-button>
              <el-button
                type="danger"
                size="small"
                @click="handleApprove(scope.row, false)"
              >
                拒绝
              </el-button>
            </div>
            <!-- 已借出/逾期状态：显示还书按钮 -->
            <el-button
              v-else-if="scope.row.status === 'borrowed' || scope.row.status === 'overdue'"
              type="success"
              size="small"
              @click="handleReturn(scope.row)"
            >
              还书
            </el-button>
            <!-- 其他状态 -->
            <span v-else style="color: #999;">-</span>
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

    <!-- 借书对话框 -->
    <el-dialog
      v-model="borrowDialogVisible"
      title="借书"
      width="500px"
    >
      <el-form
        :model="borrowForm"
        ref="borrowFormRef"
        label-width="100px"
      >
        <el-form-item label="读者ID" prop="reader_id">
          <el-input-number
            v-model="borrowForm.reader_id"
            :min="1"
            style="width: 100%;"
            placeholder="请输入读者ID"
          />
        </el-form-item>
        <el-form-item label="图书ID" prop="book_id">
          <el-input-number
            v-model="borrowForm.book_id"
            :min="1"
            style="width: 100%;"
            placeholder="请输入图书ID"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="borrowDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleBorrowSubmit">确定</el-button>
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
  name: 'BorrowList',
  components: {
    SearchIcon: Search
  },
  setup() {
    const loading = ref(false)
    const borrowList = ref([])
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(10)
    const searchKeyword = ref('')
    const statusFilter = ref('') // 状态筛选
    const borrowDialogVisible = ref(false)
    const borrowFormRef = ref(null)

    const borrowForm = reactive({
      reader_id: null,
      book_id: null
    })

    const formatDate = (dateStr) => {
      if (!dateStr) return '-'
      const date = new Date(dateStr)
      return date.toLocaleDateString('zh-CN')
    }

    const getStatusType = (status) => {
      const map = {
        pending: 'warning',
        borrowed: 'success',
        returned: 'info',
        overdue: 'danger',
        renewed: 'warning',
        rejected: 'info'
      }
      return map[status] || 'info'
    }

    const getStatusText = (status) => {
      const map = {
        pending: '待审批',
        borrowed: '已借出',
        returned: '已归还',
        overdue: '逾期',
        renewed: '已续借',
        rejected: '已拒绝'
      }
      return map[status] || status
    }

    const fetchBorrowList = async () => {
      loading.value = true
      try {
        const params = {
          page: currentPage.value,
          pageSize: pageSize.value
        }
        if (searchKeyword.value) {
          params.keyword = searchKeyword.value
        }
        // 添加状态筛选参数
        if (statusFilter.value) {
          params.status = statusFilter.value
        }

        const response = await axios.get('/borrow/getBorrowList', { params })
        if (response.code === 200) {
          borrowList.value = response.data.list || []
          total.value = response.data.total || 0
          
          // 调试：查看第一条记录的结构
          if (borrowList.value.length > 0) {
            console.log('第一条借阅记录:', borrowList.value[0])
            console.log('ID字段:', borrowList.value[0].ID, borrowList.value[0].id)
          }
        } else {
          ElMessage.error(response.msg || '获取数据失败')
        }
      } catch (error) {
        console.error('获取借阅记录失败:', error)
        ElMessage.error('获取借阅记录失败')
      } finally {
        loading.value = false
      }
    }

    const handleFilterChange = () => {
      currentPage.value = 1
      fetchBorrowList()
    }

    const handleSearch = () => {
      currentPage.value = 1
      fetchBorrowList()
    }

    const handleSizeChange = (val) => {
      pageSize.value = val
      currentPage.value = 1
      fetchBorrowList()
    }

    const handlePageChange = (val) => {
      currentPage.value = val
      fetchBorrowList()
    }

    const handleBorrow = () => {
      borrowForm.reader_id = null
      borrowForm.book_id = null
      borrowDialogVisible.value = true
    }

    const handleBorrowSubmit = async () => {
      if (!borrowForm.reader_id || !borrowForm.book_id) {
        ElMessage.warning('请填写完整信息')
        return
      }

      try {
        const response = await axios.post('/borrow/borrowBook', borrowForm)
        if (response.code === 200) {
          ElMessage.success('借书成功')
          borrowDialogVisible.value = false
          fetchBorrowList()
        } else {
          ElMessage.error(response.msg || '借书失败')
        }
      } catch (error) {
        console.error('借书失败:', error)
        ElMessage.error('借书失败')
      }
    }

    const handleApprove = (row, approved) => {
      const action = approved ? '批准' : '拒绝'
      const readerName = row.reader?.user?.real_name || row.reader?.user?.username || '该用户'
      const bookTitle = row.book?.title || '该图书'
      
      // 兼容大小写 ID 字段
      const recordId = row.ID || row.id
      console.log('点击审批按钮, row:', row, 'recordId:', recordId)
      
      if (!recordId) {
        ElMessage.error('记录ID不存在')
        return
      }
      
      if (approved) {
        ElMessageBox.confirm(`确定批准 ${readerName} 借阅《${bookTitle}》吗？`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(async () => {
          await submitApproval(recordId, approved, '')
        }).catch(() => {})
      } else {
        ElMessageBox.prompt('请输入拒绝原因', '拒绝借阅申请', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          inputPlaceholder: '请输入拒绝原因',
          inputValidator: (value) => {
            if (!value || value.trim() === '') {
              return '拒绝原因不能为空'
            }
            return true
          }
        }).then(async ({ value }) => {
          await submitApproval(recordId, approved, value)
        }).catch(() => {})
      }
    }

    const submitApproval = async (recordId, approved, rejectReason) => {
      try {
        console.log('提交审批请求:', {
          record_id: recordId,
          approved: approved,
          reject_reason: rejectReason,
          '类型检查': {
            recordId_type: typeof recordId,
            approved_type: typeof approved,
            rejectReason_type: typeof rejectReason
          }
        })
        
        const response = await axios.post('/borrow/approve', {
          record_id: recordId,
          approved: approved,
          reject_reason: rejectReason
        })
        
        console.log('审批响应:', response)
        
        if (response.code === 200) {
          ElMessage.success(approved ? '已批准借阅申请' : '已拒绝借阅申请')
          fetchBorrowList()
        } else {
          ElMessage.error(response.msg || '操作失败')
        }
      } catch (error) {
        console.error('审批失败:', error)
        ElMessage.error('操作失败，请稍后再试')
      }
    }

    const handleReturn = (row) => {
      // 兼容大小写 ID 字段
      const recordId = row.ID || row.id
      
      if (!recordId) {
        ElMessage.error('记录ID不存在')
        return
      }
      
      ElMessageBox.confirm('确定要归还该图书吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const response = await axios.post('/borrow/returnBook', {
            id: recordId
          })
          if (response.code === 200) {
            if (response.data.fine_amount > 0) {
              ElMessage.warning(`还书成功，产生逾期费用 ¥${response.data.fine_amount.toFixed(2)}`)
            } else {
              ElMessage.success('还书成功')
            }
            fetchBorrowList()
          } else {
            ElMessage.error(response.msg || '还书失败')
          }
        } catch (error) {
          console.error('还书失败:', error)
          ElMessage.error('还书失败')
        }
      }).catch(() => {})
    }

    onMounted(() => {
      fetchBorrowList()
    })

    return {
      loading,
      borrowList,
      total,
      currentPage,
      pageSize,
      searchKeyword,
      statusFilter,
      borrowDialogVisible,
      borrowForm,
      borrowFormRef,
      formatDate,
      getStatusType,
      getStatusText,
      handleSearch,
      handleFilterChange,
      handleSizeChange,
      handlePageChange,
      handleBorrow,
      handleBorrowSubmit,
      handleApprove,
      handleReturn
    }
  }
}
</script>

<style scoped>
.borrow-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
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

