<template>
  <el-container class="layout-container">
    <el-main class="main">
      <router-view />
    </el-main>
    
    <!-- 底部功能栏 -->
    <div class="bottom-bar">
      <div class="bottom-bar-content">
        <!-- 导航菜单 -->
        <el-tooltip content="图书管理" placement="top" :show-after="100" :hide-after="0">
          <button 
            class="bottom-btn" 
            :class="{ active: $route.path === '/books' && !$route.query.view }"
            @click="$router.push('/books')"
          >
            <el-icon><Reading /></el-icon>
          </button>
        </el-tooltip>
        <el-tooltip content="图书榜单" placement="top" :show-after="100" :hide-after="0">
          <button 
            class="bottom-btn"
            :class="{ active: $route.path === '/ranking' }"
            @click="$router.push('/ranking')"
          >
            <el-icon><TrendCharts /></el-icon>
          </button>
        </el-tooltip>
        <el-tooltip content="我的借阅" placement="top" :show-after="100" :hide-after="0">
          <button 
            class="bottom-btn"
            :class="{ active: $route.path === '/books' && $route.query.view === 'borrow' }"
            @click="$router.push({ path: '/books', query: { view: 'borrow' } })"
          >
            <el-icon><List /></el-icon>
          </button>
        </el-tooltip>
        <el-tooltip v-if="hasAdminOrLibrarianRole()" content="读者管理" placement="top" :show-after="100" :hide-after="0">
          <button 
            class="bottom-btn"
            :class="{ active: $route.path === '/readers' }"
            @click="$router.push('/readers')"
          >
            <el-icon><User /></el-icon>
          </button>
        </el-tooltip>
        <el-tooltip v-if="hasAdminOrLibrarianRole()" content="借还管理" placement="top" :show-after="100" :hide-after="0">
          <button 
            class="bottom-btn"
            :class="{ active: $route.path === '/borrow' }"
            @click="$router.push('/borrow')"
          >
            <el-icon><Document /></el-icon>
          </button>
        </el-tooltip>
        
        <el-tooltip v-if="hasAdminOrLibrarianRole()" content="统计查询" placement="top" :show-after="100" :hide-after="0">
          <button 
            class="bottom-btn"
            :class="{ active: $route.path === '/statistics' }"
            @click="$router.push('/statistics')"
          >
            <el-icon><DataAnalysis /></el-icon>
          </button>
        </el-tooltip>
        <el-tooltip v-if="isAdmin()" content="系统管理" placement="top" :show-after="100" :hide-after="0">
          <button 
            class="bottom-btn"
            :class="{ active: $route.path === '/system' }"
            @click="$router.push('/system')"
          >
            <el-icon><Setting /></el-icon>
          </button>
        </el-tooltip>
        
        <!-- 分隔线 -->
        <div class="bottom-divider"></div>
        
        <!-- 用户信息 -->
        <el-tooltip content="用户信息" placement="top" :show-after="100" :hide-after="0">
          <el-dropdown @command="handleUserCommand" trigger="click">
            <button class="bottom-btn user-btn">
              <el-icon><UserFilled /></el-icon>
              <span class="user-name">{{ userInfo?.real_name || userInfo?.username || '用户' }}</span>
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-tooltip>
      </div>
    </div>
  </el-container>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Reading,
  User,
  Document,
  List,
  DataAnalysis,
  Setting,
  UserFilled,
  TrendCharts
} from '@element-plus/icons-vue'
import { hasAdminOrLibrarianRole, isAdmin, getUserInfo, removeToken } from '../utils/auth'

export default {
  name: 'Layout',
  components: {
    Reading,
    User,
    Document,
    List,
    DataAnalysis,
    Setting,
    UserFilled,
    TrendCharts
  },
  setup() {
    const router = useRouter()
    const userInfo = ref(null)

    const handleUserCommand = (command) => {
      if (command === 'logout') {
        ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          removeToken()
          ElMessage.success('已退出登录')
          router.push('/login')
        }).catch(() => {})
      } else if (command === 'profile') {
        ElMessage.info('个人信息功能待实现')
      }
    }

    onMounted(() => {
      userInfo.value = getUserInfo()
    })

    return {
      userInfo,
      handleUserCommand,
      hasAdminOrLibrarianRole,
      isAdmin
    }
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.main {
  background-color: #f5f5f5;
  padding: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding-bottom: 100px; /* 为底部导航栏留出空间 */
  height: 100vh;
}

/* 底部功能栏 */
.bottom-bar {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 100;
}

.bottom-bar-content {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border-radius: 9999px;
  padding: 12px 20px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  display: flex;
  align-items: center;
  gap: 16px;
}

.bottom-btn {
  min-width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  cursor: pointer;
  color: #6b7280;
  font-size: 20px;
  padding: 0 12px;
  border-radius: 20px;
  transition: all 0.2s;
}

.bottom-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: #111827;
}

.bottom-btn.active {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.bottom-divider {
  width: 1px;
  height: 24px;
  background: #e5e7eb;
  margin: 0 8px;
}

.user-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  white-space: nowrap;
}

/* 自定义tooltip样式 */
:deep(.el-tooltip__popper) {
  font-size: 16px !important;
  padding: 10px 16px !important;
  font-weight: 500 !important;
}
</style>
