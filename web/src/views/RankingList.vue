<template>
  <div class="ranking-page">
    <div class="ranking-shell">
      <div class="ranking-header">
        <div class="header-main">
          <h1 class="ranking-title">图书榜单</h1>
          <p class="ranking-subtitle">按点赞与收藏热度，查看本周/本月最受欢迎图书</p>
        </div>

        <div class="ranking-tabs">
          <button
            class="tab-btn"
            :class="{ active: currentType === 'like' }"
            @click="switchType('like')"
          >
            点赞榜
          </button>
          <button
            class="tab-btn"
            :class="{ active: currentType === 'favorite' }"
            @click="switchType('favorite')"
          >
            收藏榜
          </button>
        </div>

        <div class="period-tabs">
          <button
            class="period-btn"
            :class="{ active: currentPeriod === 'week' }"
            @click="switchPeriod('week')"
          >
            周榜
          </button>
          <button
            class="period-btn"
            :class="{ active: currentPeriod === 'month' }"
            @click="switchPeriod('month')"
          >
            月榜
          </button>
        </div>
      </div>

      <div v-if="rankingData" class="ranking-info">
        <div class="info-item">
          <span class="info-label">榜单周期</span>
          <span class="info-value">{{ rankingData.period_key }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">更新时间</span>
          <span class="info-value">{{ formatTime(rankingData.updated_at) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">上榜图书</span>
          <span class="info-value">{{ rankingData.total }} 本</span>
        </div>
      </div>

      <div v-if="loading" class="loading-container">
        <el-icon class="is-loading"><Loading /></el-icon>
        <p>加载榜单中...</p>
      </div>

      <div v-else-if="rankingData && rankingData.items && rankingData.items.length > 0" class="ranking-list">
        <div
          v-for="item in rankingData.items"
          :key="item.book_id"
          class="ranking-item"
          :class="{
            'rank-1': item.rank === 1,
            'rank-2': item.rank === 2,
            'rank-3': item.rank === 3
          }"
        >
          <div class="rank-badge">
            <span class="rank-number">{{ String(item.rank).padStart(2, '0') }}</span>
          </div>

          <div class="book-cover">
            <img
              v-if="item.book && item.book.cover_image"
              :src="item.book.cover_image"
              :alt="item.book.title"
              referrerpolicy="no-referrer"
              @error="handleImageError"
            />
            <div v-else class="cover-placeholder">
              <span>BOOK</span>
            </div>
          </div>

          <div class="book-info">
            <h3 class="book-title">{{ item.book ? item.book.title : `图书 ID: ${item.book_id}` }}</h3>
            <p v-if="item.book && item.book.author" class="book-author">{{ item.book.author }}</p>
            <div v-if="item.book && item.book.categories && item.book.categories.length > 0" class="book-categories">
              <span
                v-for="cat in item.book.categories.slice(0, 3)"
                :key="cat.id"
                class="category-tag"
              >
                {{ cat.name }}
              </span>
            </div>
          </div>

          <div class="score-badge">
            <span class="score-label">{{ currentType === 'like' ? 'LIKES' : 'FAV' }}</span>
            <span class="score-value">{{ item.score }}</span>
          </div>
        </div>
      </div>

      <div v-else class="empty-state">
        <el-icon class="empty-icon"><Document /></el-icon>
        <p>暂无榜单数据</p>
        <p class="empty-hint">{{ getEmptyHint() }}</p>
      </div>

      <div v-if="hasAdminRole()" class="admin-actions">
        <el-button
          type="warning"
          :loading="rebuilding"
          @click="handleRebuild"
        >
          重建当前榜单
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading, Document } from '@element-plus/icons-vue'
import { getRanking, rebuildRanking } from '@/api/ranking'
import { isAdmin } from '@/utils/auth'

export default {
  name: 'RankingList',
  components: {
    Loading,
    Document
  },
  setup() {
    const loading = ref(false)
    const rebuilding = ref(false)
    const currentType = ref('like') // 'like' 或 'favorite'
    const currentPeriod = ref('week') // 'week' 或 'month'
    const rankingData = ref(null)

    // 检查是否是管理员
    const hasAdminRole = () => {
      return isAdmin()
    }

    // 获取榜单数据
    const fetchRanking = async () => {
      loading.value = true
      try {
        console.log('获取榜单:', { type: currentType.value, period: currentPeriod.value })
        const response = await getRanking(currentType.value, currentPeriod.value, 100)
        console.log('榜单响应:', response)
        
        if (response.code === 200 && response.data) {
          rankingData.value = response.data
          ElMessage.success('榜单加载成功')
        } else {
          ElMessage.error(response.msg || '获取榜单失败')
          rankingData.value = null
        }
      } catch (error) {
        console.error('获取榜单失败:', error)
        ElMessage.error('获取榜单失败，请稍后再试')
        rankingData.value = null
      } finally {
        loading.value = false
      }
    }

    // 切换榜单类型
    const switchType = (type) => {
      if (currentType.value === type) return
      currentType.value = type
      fetchRanking()
    }

    // 切换榜单周期
    const switchPeriod = (period) => {
      if (currentPeriod.value === period) return
      currentPeriod.value = period
      fetchRanking()
    }

    // 重建榜单（管理员功能）
    const handleRebuild = async () => {
      try {
        await ElMessageBox.confirm(
          `确定要重建当前榜单吗？这将重新计算所有数据。`,
          '重建榜单',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        rebuilding.value = true
        const response = await rebuildRanking(currentType.value, currentPeriod.value)
        
        if (response.code === 200) {
          ElMessage.success('榜单重建成功')
          // 重新获取榜单
          await fetchRanking()
        } else {
          ElMessage.error(response.msg || '榜单重建失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('重建榜单失败:', error)
          ElMessage.error('重建榜单失败，请稍后再试')
        }
      } finally {
        rebuilding.value = false
      }
    }

    // 格式化时间
    const formatTime = (timeStr) => {
      if (!timeStr) return '-'
      try {
        const date = new Date(timeStr)
        return date.toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit'
        })
      } catch {
        return timeStr
      }
    }

    // 图片加载错误处理
    const handleImageError = (e) => {
      if (e.target.dataset.errorHandled) {
        e.target.style.display = 'none'
        return
      }
      e.target.dataset.errorHandled = 'true'
      e.target.style.display = 'none'
    }

    // 获取空状态提示
    const getEmptyHint = () => {
      const typeText = currentType.value === 'like' ? '点赞' : '收藏'
      const periodText = currentPeriod.value === 'week' ? '本周' : '本月'
      return `${periodText}还没有图书${typeText}数据，快去${typeText}你喜欢的图书吧！`
    }

    onMounted(() => {
      fetchRanking()
    })

    return {
      loading,
      rebuilding,
      currentType,
      currentPeriod,
      rankingData,
      hasAdminRole,
      switchType,
      switchPeriod,
      handleRebuild,
      formatTime,
      handleImageError,
      getEmptyHint
    }
  }
}
</script>

<style scoped>
.ranking-page {
  --rank-bg: #f5f7fb;
  --rank-panel-bg: rgba(255, 255, 255, 0.92);
  --rank-panel-border: rgba(15, 23, 42, 0.08);
  --rank-shadow: 0 10px 30px rgba(15, 23, 42, 0.08);
  --rank-text: #0f172a;
  --rank-muted: #64748b;
  --rank-soft: #e2e8f0;
  --rank-accent: #3b82f6;
  --rank-accent-soft: rgba(59, 130, 246, 0.14);
  min-height: 100vh;
  background: var(--rank-bg);
  padding: 24px 20px 120px;
}

.ranking-shell {
  width: min(1120px, 100%);
  margin: 0 auto;
}

.ranking-header {
  background: var(--rank-panel-bg);
  border: 1px solid var(--rank-panel-border);
  box-shadow: var(--rank-shadow);
  border-radius: 16px;
  padding: 22px;
  display: grid;
  grid-template-columns: 1fr auto auto;
  align-items: center;
  gap: 14px;
  margin-bottom: 14px;
}

.header-main {
  min-width: 0;
}

.ranking-title {
  margin: 0;
  font-size: 26px;
  line-height: 1.2;
  color: var(--rank-text);
  letter-spacing: 0.02em;
}

.ranking-subtitle {
  margin: 6px 0 0;
  color: var(--rank-muted);
  font-size: 13px;
}

.ranking-tabs {
  display: flex;
  gap: 8px;
  background: var(--rank-soft);
  padding: 4px;
  border-radius: 999px;
}

.tab-btn {
  padding: 8px 14px;
  font-size: 13px;
  font-weight: 600;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--rank-muted);
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.tab-btn.active {
  background: var(--rank-panel-bg);
  color: var(--rank-text);
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.12);
}

.period-tabs {
  display: flex;
  gap: 8px;
  background: var(--rank-soft);
  padding: 4px;
  border-radius: 999px;
}

.period-btn {
  padding: 8px 14px;
  font-size: 13px;
  font-weight: 500;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--rank-muted);
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.period-btn.active {
  background: var(--rank-accent-soft);
  color: var(--rank-accent);
}

.ranking-info {
  background: var(--rank-panel-bg);
  border: 1px solid var(--rank-panel-border);
  box-shadow: var(--rank-shadow);
  border-radius: 14px;
  padding: 14px 16px;
  margin-bottom: 14px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.info-item {
  flex: 1;
  min-width: 180px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  padding: 6px 8px;
  border-radius: 10px;
  background: rgba(148, 163, 184, 0.08);
}

.info-label {
  color: var(--rank-muted);
  font-size: 12px;
  font-weight: 500;
}

.info-value {
  color: var(--rank-text);
  font-weight: 600;
  font-size: 14px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  background: var(--rank-panel-bg);
  border: 1px solid var(--rank-panel-border);
  border-radius: 16px;
  color: var(--rank-accent);
}

.loading-container .el-icon {
  font-size: 36px;
  margin-bottom: 10px;
}

.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 14px;
  background: var(--rank-panel-bg);
  border: 1px solid var(--rank-panel-border);
  border-radius: 14px;
  padding: 14px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.ranking-item:hover {
  transform: translateY(-1px);
  box-shadow: var(--rank-shadow);
}

.ranking-item.rank-1 {
  border-color: rgba(245, 158, 11, 0.45);
}

.ranking-item.rank-2 {
  border-color: rgba(148, 163, 184, 0.55);
}

.ranking-item.rank-3 {
  border-color: rgba(180, 83, 9, 0.4);
}

.rank-badge {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: rgba(148, 163, 184, 0.16);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.rank-number {
  color: var(--rank-accent);
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.book-cover {
  flex-shrink: 0;
  width: 64px;
  height: 94px;
  border-radius: 8px;
  overflow: hidden;
  background: rgba(148, 163, 184, 0.16);
}

.book-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--rank-muted);
  font-size: 10px;
  letter-spacing: 0.08em;
  font-weight: 700;
}

.book-info {
  flex: 1;
  min-width: 0;
}

.book-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--rank-text);
  margin: 0 0 4px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.book-author {
  font-size: 13px;
  color: var(--rank-muted);
  margin: 0 0 6px 0;
}

.book-categories {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.category-tag {
  padding: 2px 8px;
  background: rgba(148, 163, 184, 0.14);
  color: var(--rank-muted);
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
}

.score-badge {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  min-width: 72px;
  padding: 8px 10px;
  background: var(--rank-accent-soft);
  border-radius: 10px;
  color: var(--rank-accent);
}

.score-label {
  font-size: 10px;
  letter-spacing: 0.08em;
  opacity: 0.9;
}

.score-value {
  font-size: 22px;
  font-weight: bold;
  line-height: 1;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 70px 20px;
  background: var(--rank-panel-bg);
  border: 1px solid var(--rank-panel-border);
  border-radius: 16px;
  color: var(--rank-muted);
}

.empty-icon {
  font-size: 52px;
  margin-bottom: 12px;
  color: rgba(148, 163, 184, 0.8);
}

.empty-state p {
  font-size: 16px;
  margin: 5px 0;
}

.empty-hint {
  color: var(--rank-muted);
  opacity: 0.8;
  font-size: 13px;
}

.admin-actions {
  display: flex;
  justify-content: center;
  margin-top: 12px;
  padding: 12px;
  background: var(--rank-panel-bg);
  border: 1px solid var(--rank-panel-border);
  border-radius: 14px;
}

:global(body.app-theme-dark .ranking-page) {
  --rank-bg: #0f172a;
  --rank-panel-bg: rgba(15, 23, 42, 0.88);
  --rank-panel-border: rgba(148, 163, 184, 0.24);
  --rank-shadow: 0 10px 30px rgba(2, 6, 23, 0.55);
  --rank-text: #f8fafc;
  --rank-muted: #cbd5e1;
  --rank-soft: rgba(148, 163, 184, 0.18);
  --rank-accent: #93c5fd;
  --rank-accent-soft: rgba(59, 130, 246, 0.24);
}

@media (max-width: 768px) {
  .ranking-page {
    padding: 16px 12px 110px;
  }

  .ranking-header {
    grid-template-columns: 1fr;
    padding: 16px;
    gap: 10px;
  }

  .ranking-title {
    font-size: 24px;
  }

  .ranking-item {
    gap: 10px;
    padding: 12px;
  }
}
</style>

