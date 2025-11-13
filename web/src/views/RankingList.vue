<template>
  <div class="ranking-page">
    <!-- 顶部导航栏 -->
    <div class="ranking-header">
      <h1 class="ranking-title">📊 图书榜单</h1>
      
      <!-- 榜单类型切换 -->
      <div class="ranking-tabs">
        <button 
          class="tab-btn"
          :class="{ active: currentType === 'like' }"
          @click="switchType('like')"
        >
          ❤️ 点赞榜
        </button>
        <button 
          class="tab-btn"
          :class="{ active: currentType === 'favorite' }"
          @click="switchType('favorite')"
        >
          ⭐ 收藏榜
        </button>
      </div>
      
      <!-- 周期切换 -->
      <div class="period-tabs">
        <button 
          class="period-btn"
          :class="{ active: currentPeriod === 'week' }"
          @click="switchPeriod('week')"
        >
          📅 周榜
        </button>
        <button 
          class="period-btn"
          :class="{ active: currentPeriod === 'month' }"
          @click="switchPeriod('month')"
        >
          📆 月榜
        </button>
      </div>
    </div>

    <!-- 榜单信息 -->
    <div v-if="rankingData" class="ranking-info">
      <div class="info-item">
        <span class="info-label">榜单周期：</span>
        <span class="info-value">{{ rankingData.period_key }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">更新时间：</span>
        <span class="info-value">{{ formatTime(rankingData.updated_at) }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">上榜图书：</span>
        <span class="info-value">{{ rankingData.total }} 本</span>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-icon class="is-loading"><Loading /></el-icon>
      <p>加载榜单中...</p>
    </div>

    <!-- 榜单列表 -->
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
        <!-- 排名 -->
        <div class="rank-badge">
          <span v-if="item.rank === 1" class="medal gold">🥇</span>
          <span v-else-if="item.rank === 2" class="medal silver">🥈</span>
          <span v-else-if="item.rank === 3" class="medal bronze">🥉</span>
          <span v-else class="rank-number">{{ item.rank }}</span>
        </div>

        <!-- 图书封面 -->
        <div class="book-cover">
          <img 
            v-if="item.book && item.book.cover_image"
            :src="item.book.cover_image"
            :alt="item.book.title"
            referrerpolicy="no-referrer"
            @error="handleImageError"
          />
          <div v-else class="cover-placeholder">
            <span>📚</span>
          </div>
        </div>

        <!-- 图书信息 -->
        <div class="book-info">
          <h3 class="book-title">{{ item.book ? item.book.title : `图书 ID: ${item.book_id}` }}</h3>
          <p v-if="item.book && item.book.author" class="book-author">作者：{{ item.book.author }}</p>
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

        <!-- 分数 -->
        <div class="score-badge">
          <span class="score-icon">{{ currentType === 'like' ? '❤️' : '⭐' }}</span>
          <span class="score-value">{{ item.score }}</span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <el-icon class="empty-icon"><Document /></el-icon>
      <p>暂无榜单数据</p>
      <p class="empty-hint">{{ getEmptyHint() }}</p>
    </div>

    <!-- 管理员操作：重建榜单 -->
    <div v-if="hasAdminRole()" class="admin-actions">
      <el-button 
        type="warning" 
        :loading="rebuilding"
        @click="handleRebuild"
      >
        🔄 重建当前榜单
      </el-button>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading, Document } from '@element-plus/icons-vue'
import { getRanking, rebuildRanking } from '@/api/ranking'
import { hasAdminOrLibrarianRole } from '@/utils/auth'

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
      return hasAdminOrLibrarianRole()
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
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

/* 顶部导航栏 */
.ranking-header {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 30px;
  margin-bottom: 20px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.ranking-title {
  font-size: 32px;
  font-weight: bold;
  color: #2d3748;
  margin: 0 0 20px 0;
  text-align: center;
}

/* 榜单类型切换 */
.ranking-tabs {
  display: flex;
  gap: 15px;
  justify-content: center;
  margin-bottom: 20px;
}

.tab-btn {
  flex: 1;
  max-width: 200px;
  padding: 15px 30px;
  font-size: 18px;
  font-weight: 600;
  border: 3px solid #e2e8f0;
  border-radius: 15px;
  background: white;
  color: #64748b;
  cursor: pointer;
  transition: all 0.3s ease;
}

.tab-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.tab-btn.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-color: transparent;
  box-shadow: 0 4px 20px rgba(102, 126, 234, 0.4);
}

/* 周期切换 */
.period-tabs {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.period-btn {
  padding: 10px 25px;
  font-size: 16px;
  font-weight: 500;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  background: white;
  color: #64748b;
  cursor: pointer;
  transition: all 0.3s ease;
}

.period-btn:hover {
  border-color: #667eea;
  color: #667eea;
}

.period-btn.active {
  background: #667eea;
  color: white;
  border-color: #667eea;
}

/* 榜单信息 */
.ranking-info {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 15px;
  padding: 20px;
  margin-bottom: 20px;
  display: flex;
  gap: 30px;
  justify-content: center;
  flex-wrap: wrap;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-label {
  color: #64748b;
  font-weight: 500;
}

.info-value {
  color: #2d3748;
  font-weight: 600;
}

/* 加载状态 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 20px;
  color: #667eea;
}

.loading-container .el-icon {
  font-size: 48px;
  margin-bottom: 15px;
}

/* 榜单列表 */
.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 20px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 15px;
  padding: 20px;
  transition: all 0.3s ease;
  border: 2px solid transparent;
}

.ranking-item:hover {
  transform: translateX(5px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.ranking-item.rank-1 {
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.2), rgba(255, 223, 0, 0.1));
  border-color: #ffd700;
}

.ranking-item.rank-2 {
  background: linear-gradient(135deg, rgba(192, 192, 192, 0.2), rgba(192, 192, 192, 0.1));
  border-color: #c0c0c0;
}

.ranking-item.rank-3 {
  background: linear-gradient(135deg, rgba(205, 127, 50, 0.2), rgba(205, 127, 50, 0.1));
  border-color: #cd7f32;
}

/* 排名徽章 */
.rank-badge {
  flex-shrink: 0;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36px;
  font-weight: bold;
}

.medal {
  font-size: 48px;
}

.rank-number {
  color: #667eea;
  font-size: 28px;
}

/* 图书封面 */
.book-cover {
  flex-shrink: 0;
  width: 80px;
  height: 120px;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
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
  background: linear-gradient(135deg, #667eea, #764ba2);
  font-size: 36px;
}

/* 图书信息 */
.book-info {
  flex: 1;
  min-width: 0;
}

.book-title {
  font-size: 20px;
  font-weight: 600;
  color: #2d3748;
  margin: 0 0 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.book-author {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 8px 0;
}

.book-categories {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.category-tag {
  padding: 4px 12px;
  background: #e0e7ff;
  color: #667eea;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

/* 分数徽章 */
.score-badge {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  padding: 15px 20px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-radius: 12px;
  color: white;
}

.score-icon {
  font-size: 24px;
}

.score-value {
  font-size: 24px;
  font-weight: bold;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 20px;
  color: #64748b;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 20px;
  color: #cbd5e0;
}

.empty-state p {
  font-size: 18px;
  margin: 5px 0;
}

.empty-hint {
  color: #a0aec0;
  font-size: 14px;
}

/* 管理员操作 */
.admin-actions {
  display: flex;
  justify-content: center;
  margin-top: 20px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 15px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .ranking-header {
    padding: 20px;
  }

  .ranking-title {
    font-size: 24px;
  }

  .ranking-tabs {
    flex-direction: column;
  }

  .tab-btn {
    max-width: none;
  }

  .ranking-item {
    flex-direction: column;
    text-align: center;
  }

  .rank-badge {
    order: -1;
  }

  .book-info {
    order: 1;
  }

  .score-badge {
    order: 2;
  }
}
</style>

