<template>
  <div class="my-borrow-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ isAdmin ? '借阅管理' : '我的借阅' }} <small style="color: #999;">(管理员: {{ isAdmin }})</small></span>
          <div v-if="isAdmin" class="header-actions">
            <el-radio-group v-model="statusFilter" @change="handleFilterChange">
              <el-radio-button label="">全部</el-radio-button>
              <el-radio-button label="pending">待审批</el-radio-button>
              <el-radio-button label="borrowed">已借出</el-radio-button>
              <el-radio-button label="overdue">逾期</el-radio-button>
              <el-radio-button label="returned">已归还</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>

      <!-- 借阅记录表格 -->
      <el-table
        :data="borrowList"
        style="width: 100%; margin-top: 20px;"
        v-loading="loading"
        border
      >
        <el-table-column v-if="isAdmin" prop="reader.user.username" label="借阅人" width="120" />
        <el-table-column prop="book.title" label="书名" min-width="150" />
        <el-table-column prop="book.author" label="作者" width="120" />
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
        <el-table-column prop="status" label="状态" width="150">
          <template #default="scope">
            <div style="display: flex; flex-direction: column; align-items: center; gap: 4px;">
              <el-tag :type="getStatusType(scope.row.status)">
                {{ getStatusText(scope.row.status) }}
              </el-tag>
              <small style="color: #999; font-size: 11px;">({{ scope.row.status }})</small>
            </div>
          </template>
        </el-table-column>
        <el-table-column v-if="!isAdmin" prop="renew_count" label="续借次数" width="100" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <!-- 调试信息 -->
            <div style="font-size: 10px; color: #999; margin-bottom: 4px;">
              状态: {{ scope.row.status }} | 管理员: {{ isAdmin }}
            </div>
            
            <!-- 待审批状态：显示批准/拒绝按钮（管理员） -->
            <div v-if="scope.row.status === 'pending' && isAdmin" style="display: flex; gap: 8px; justify-content: center;">
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
            <!-- 已借出状态：显示续借按钮（非管理员且未达续借次数上限） -->
            <el-button
              v-else-if="!isAdmin && scope.row.status === 'borrowed' && scope.row.renew_count < 2"
              type="primary"
              size="small"
              @click="handleRenew(scope.row)"
            >
              续借
            </el-button>
            <!-- 已达最大续借次数 -->
            <span v-else-if="!isAdmin && scope.row.renew_count >= 2" style="color: #999;">
              已达最大续借次数
            </span>
            <!-- 其他状态 -->
            <span v-else style="color: #999;">调试: status={{ scope.row.status }}, isAdmin={{ isAdmin }}</span>
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
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'
import { hasAdminOrLibrarianRole } from '@/utils/auth'

export default {
  name: 'MyBorrowList',
  setup() {
    const loading = ref(false)
    const borrowList = ref([])
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(10)
    const isAdmin = ref(hasAdminOrLibrarianRole())
    const statusFilter = ref('') // 状态筛选

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
      console.log('=== fetchBorrowList 开始 ===')
      console.log('isAdmin.value:', isAdmin.value)
      
      loading.value = true
      try {
        const params = {
          page: currentPage.value,
          pageSize: pageSize.value
        }
        
        // 如果有状态筛选，添加到参数中
        if (statusFilter.value) {
          params.status = statusFilter.value
        }

        // 管理员调用管理接口，普通用户调用个人接口
        const apiUrl = isAdmin.value ? '/borrow/getBorrowList' : '/borrow/getMyBorrowList'
        console.log('API URL:', apiUrl)
        console.log('请求参数:', params)
        
        const response = await axios.get(apiUrl, { params })
        console.log('API 响应:', response)
        
        if (response.code === 200) {
          borrowList.value = response.data.list || []
          total.value = response.data.total || 0
          
          console.log('✅ 数据加载成功')
          console.log('borrowList 长度:', borrowList.value.length)
          console.log('borrowList 数据:', borrowList.value)
          
          if (borrowList.value.length > 0) {
            console.log('第一条记录:', borrowList.value[0])
            console.log('第一条记录状态:', borrowList.value[0].status)
            console.log('状态类型:', typeof borrowList.value[0].status)
          }
        } else {
          console.log('❌ API 返回非 200:', response)
          ElMessage.error(response.msg || '获取数据失败')
        }
      } catch (error) {
        console.error('❌ 获取借阅记录失败:', error)
        ElMessage.error('获取借阅记录失败')
      } finally {
        loading.value = false
        console.log('=== fetchBorrowList 结束 ===')
      }
    }

    const handleFilterChange = () => {
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

    const handleApprove = (row, approved) => {
      const action = approved ? '批准' : '拒绝'
      const content = approved ? `确定批准 ${row.reader?.user?.username || '该用户'} 借阅《${row.book?.title}》吗？` : '请输入拒绝原因'
      
      if (approved) {
        ElMessageBox.confirm(content, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(async () => {
          await submitApproval(row.id, approved, '')
        }).catch(() => {})
      } else {
        ElMessageBox.prompt(content, '拒绝借阅申请', {
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
          await submitApproval(row.id, approved, value)
        }).catch(() => {})
      }
    }

    const submitApproval = async (recordId, approved, rejectReason) => {
      try {
        const response = await axios.post('/borrow/approve', {
          record_id: recordId,
          approved: approved,
          reject_reason: rejectReason
        })
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

    const handleRenew = (row) => {
      ElMessageBox.confirm('确定要续借该图书吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const response = await axios.post('/borrow/renewBook', {
            record_id: row.id
          })
          if (response.code === 200) {
            ElMessage.success('续借成功')
            fetchBorrowList()
          } else {
            ElMessage.error(response.msg || '续借失败')
          }
        } catch (error) {
          console.error('续借失败:', error)
          ElMessage.error('续借失败')
        }
      }).catch(() => {})
    }

    onMounted(() => {
      console.log('=== MyBorrowList 组件挂载 ===')
      console.log('isAdmin:', isAdmin.value)
      console.log('hasAdminOrLibrarianRole():', hasAdminOrLibrarianRole())
      fetchBorrowList()
    })

    return {
      loading,
      borrowList,
      total,
      currentPage,
      pageSize,
      isAdmin,
      statusFilter,
      formatDate,
      getStatusType,
      getStatusText,
      handleFilterChange,
      handleSizeChange,
      handlePageChange,
      handleApprove,
      handleRenew
    }
  }
}
</script>

<style scoped>
.my-borrow-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header span {
  font-size: 18px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

