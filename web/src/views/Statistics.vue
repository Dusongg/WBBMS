<template>
  <div class="statistics">
    <el-row :gutter="20">
      <!-- 统计卡片 -->
      <el-col :span="6" v-for="(stat, index) in statistics" :key="index">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stat.value }}</div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 热门图书 -->
    <el-card style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>热门图书（借阅次数最多）</span>
        </div>
      </template>

      <el-table
        :data="popularBooks"
        style="width: 100%;"
        v-loading="loading"
        border
      >
        <el-table-column type="index" label="排名" width="80" />
        <el-table-column prop="title" label="书名" min-width="200" />
        <el-table-column prop="author" label="作者" width="120" />
        <el-table-column prop="borrow_count" label="借阅次数" width="120" />
      </el-table>
    </el-card>

    <!-- 借阅统计 -->
    <el-card style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>借阅统计</span>
          <div>
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              @change="handleDateChange"
              style="width: 300px;"
            />
            <el-button type="primary" @click="fetchBorrowStatistics" style="margin-left: 10px;">
              查询
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="borrowStatistics"
        style="width: 100%;"
        v-loading="statisticsLoading"
        border
      >
        <el-table-column prop="reader.user.real_name" label="读者" width="120" />
        <el-table-column prop="book.title" label="书名" min-width="150" />
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
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'

export default {
  name: 'Statistics',
  setup() {
    const loading = ref(false)
    const statisticsLoading = ref(false)
    const statistics = ref([])
    const popularBooks = ref([])
    const borrowStatistics = ref([])
    const dateRange = ref(null)

    const formatDate = (dateStr) => {
      if (!dateStr) return '-'
      const date = new Date(dateStr)
      return date.toLocaleDateString('zh-CN')
    }

    const getStatusType = (status) => {
      const map = {
        borrowed: 'success',
        returned: 'info',
        overdue: 'danger',
        renewed: 'warning'
      }
      return map[status] || 'info'
    }

    const getStatusText = (status) => {
      const map = {
        borrowed: '已借出',
        returned: '已归还',
        overdue: '逾期',
        renewed: '已续借'
      }
      return map[status] || status
    }

    const fetchStatistics = async () => {
      loading.value = true
      try {
        const response = await axios.get('/statistics/getStatistics')
        if (response.code === 200) {
          const data = response.data
          statistics.value = [
            { label: '图书总数', value: data.total_books || 0 },
            { label: '可借图书', value: data.available_books || 0 },
            { label: '读者总数', value: data.total_readers || 0 },
            { label: '借阅中', value: data.borrowing_count || 0 },
            { label: '逾期数量', value: data.overdue_count || 0 },
            { label: '本月借阅', value: data.month_borrow_count || 0 },
            { label: '本月归还', value: data.month_return_count || 0 }
          ]
        } else {
          ElMessage.error(response.msg || '获取统计信息失败')
        }
      } catch (error) {
        console.error('获取统计信息失败:', error)
        ElMessage.error('获取统计信息失败')
      } finally {
        loading.value = false
      }
    }

    const fetchPopularBooks = async () => {
      try {
        const response = await axios.get('/statistics/getPopularBooks')
        if (response.code === 200) {
          popularBooks.value = response.data || []
        } else {
          ElMessage.error(response.msg || '获取热门图书失败')
        }
      } catch (error) {
        console.error('获取热门图书失败:', error)
        ElMessage.error('获取热门图书失败')
      }
    }

    const fetchBorrowStatistics = async () => {
      statisticsLoading.value = true
      try {
        const params = {}
        if (dateRange.value && dateRange.value.length === 2) {
          params.start_date = formatDate(dateRange.value[0])
          params.end_date = formatDate(dateRange.value[1])
        }

        const response = await axios.get('/statistics/getBorrowStatistics', { params })
        if (response.code === 200) {
          borrowStatistics.value = response.data || []
        } else {
          ElMessage.error(response.msg || '获取借阅统计失败')
        }
      } catch (error) {
        console.error('获取借阅统计失败:', error)
        ElMessage.error('获取借阅统计失败')
      } finally {
        statisticsLoading.value = false
      }
    }

    const handleDateChange = () => {
      // 日期改变时自动查询
      if (dateRange.value && dateRange.value.length === 2) {
        fetchBorrowStatistics()
      }
    }

    onMounted(() => {
      fetchStatistics()
      fetchPopularBooks()
      fetchBorrowStatistics()
    })

    return {
      loading,
      statisticsLoading,
      statistics,
      popularBooks,
      borrowStatistics,
      dateRange,
      formatDate,
      getStatusType,
      getStatusText,
      fetchBorrowStatistics,
      handleDateChange
    }
  }
}
</script>

<style scoped>
.statistics {
  padding: 20px;
}

.stat-card {
  margin-bottom: 20px;
}

.stat-content {
  text-align: center;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

