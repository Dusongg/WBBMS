<template>
  <div class="statistics">
    <section class="stats-hero">
      <p class="hero-kicker">Dashboard</p>
      <h1 class="hero-title">借阅统计中心</h1>
      <p class="hero-desc">聚焦借阅动态、热门图书与读者行为，快速查看馆藏运营状态。</p>
    </section>

    <div class="stat-grid">
      <div class="stat-card" v-for="(stat, index) in statistics" :key="index">
        <div class="stat-content">
          <div class="stat-value">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
      </div>
    </div>

    <el-card class="panel-card">
      <template #header>
        <div class="card-header">
          <span>热门图书（借阅次数最多）</span>
        </div>
      </template>

      <el-table
        :data="popularBooks"
        class="stats-table"
        v-loading="loading"
        border
      >
        <el-table-column type="index" label="排名" width="80" />
        <el-table-column prop="title" label="书名" min-width="200" />
        <el-table-column prop="author" label="作者" width="120" />
        <el-table-column prop="borrow_count" label="借阅次数" width="120" />
      </el-table>
    </el-card>

    <el-card class="panel-card">
      <template #header>
        <div class="card-header">
          <span>借阅记录统计</span>
          <div class="toolbar">
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              @change="handleDateChange"
              class="date-picker"
            />
            <el-button type="primary" @click="fetchBorrowStatistics">
              查询
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="borrowStatistics"
        class="stats-table"
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

    const formatDateParam = (dateInput) => {
      if (!dateInput) return ''
      const date = new Date(dateInput)
      if (Number.isNaN(date.getTime())) return ''
      const y = date.getFullYear()
      const m = String(date.getMonth() + 1).padStart(2, '0')
      const d = String(date.getDate()).padStart(2, '0')
      return `${y}-${m}-${d}`
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
          params.start_date = formatDateParam(dateRange.value[0])
          params.end_date = formatDateParam(dateRange.value[1])
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
  padding: 18px;
  position: relative;
}

.statistics::before {
  content: '';
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background:
    radial-gradient(circle at 0% 0%, rgba(59, 130, 246, 0.12), transparent 45%),
    radial-gradient(circle at 100% 100%, rgba(14, 165, 233, 0.14), transparent 43%),
    linear-gradient(180deg, rgba(248, 250, 252, 0.55), rgba(239, 246, 255, 0.4));
}

.statistics > * {
  position: relative;
  z-index: 1;
}

.stats-hero {
  margin-bottom: 18px;
}

.hero-kicker {
  margin: 0;
  font-size: 11px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: #64748b;
}

.hero-title {
  margin: 6px 0 4px;
  font-size: 28px;
  line-height: 1.1;
  font-weight: 800;
  color: #0f172a;
}

.hero-desc {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
  gap: 12px;
  margin-bottom: 18px;
}

.stat-card {
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.52);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  box-shadow:
    0 10px 24px rgba(15, 23, 42, 0.1),
    0 18px 36px rgba(15, 23, 42, 0.08);
  padding: 14px 12px;
}

.stat-content {
  text-align: center;
}

.stat-value {
  font-size: 32px;
  font-weight: 800;
  color: #1d4ed8;
  margin-bottom: 6px;
}

.stat-label {
  font-size: 12px;
  color: #64748b;
}

.panel-card {
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.54);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  box-shadow:
    0 12px 28px rgba(15, 23, 42, 0.1),
    0 22px 42px rgba(15, 23, 42, 0.08);
  margin-bottom: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-weight: 700;
  color: #0f172a;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.date-picker {
  width: 300px;
}

.stats-table {
  width: 100%;
}

.stats-table :deep(.el-table__header-wrapper th),
.stats-table :deep(.el-table__body-wrapper td) {
  background: rgba(255, 255, 255, 0.5);
}

.stats-table :deep(.el-table__row:hover > td) {
  background: rgba(59, 130, 246, 0.08) !important;
}

:global(body.app-theme-dark) .hero-title,
:global(.layout-container.theme-dark) .hero-title {
  color: #f8fafc;
}

:global(body.app-theme-dark) .hero-kicker,
:global(body.app-theme-dark) .hero-desc,
:global(.layout-container.theme-dark) .hero-kicker,
:global(.layout-container.theme-dark) .hero-desc {
  color: #cbd5e1;
}

:global(body.app-theme-dark) .stat-card,
:global(body.app-theme-dark) .panel-card,
:global(.layout-container.theme-dark) .stat-card,
:global(.layout-container.theme-dark) .panel-card {
  background: rgba(15, 23, 42, 0.52);
  border-color: rgba(148, 163, 184, 0.24);
  box-shadow: 0 18px 40px rgba(2, 6, 23, 0.48);
  backdrop-filter: none;
  -webkit-backdrop-filter: none;
}

:global(body.app-theme-dark) .stat-value,
:global(.layout-container.theme-dark) .stat-value {
  color: #93c5fd;
}

:global(body.app-theme-dark) .stat-label,
:global(body.app-theme-dark) .card-header,
:global(.layout-container.theme-dark) .stat-label,
:global(.layout-container.theme-dark) .card-header {
  color: #e2e8f0;
}

:global(body.app-theme-dark) .stats-table .el-table__header-wrapper th,
:global(body.app-theme-dark) .stats-table .el-table__body-wrapper td,
:global(.layout-container.theme-dark) .stats-table .el-table__header-wrapper th,
:global(.layout-container.theme-dark) .stats-table .el-table__body-wrapper td {
  background: rgba(15, 23, 42, 0.52);
  color: #e2e8f0;
}
</style>

