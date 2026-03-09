<template>
  <el-container class="layout-container" :class="`theme-${themeMode}`">
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
        <el-tooltip content="我喜欢的" placement="top" :show-after="100" :hide-after="0">
          <button
            class="bottom-btn"
            :class="{ active: $route.path === '/books' && $route.query.view === 'like' }"
            @click="$router.push({ path: '/books', query: { view: 'like' } })"
          >
            <el-icon><Star /></el-icon>
          </button>
        </el-tooltip>
        <el-tooltip content="我收藏的" placement="top" :show-after="100" :hide-after="0">
          <button
            class="bottom-btn"
            :class="{ active: $route.path === '/books' && $route.query.view === 'favorite' }"
            @click="$router.push({ path: '/books', query: { view: 'favorite' } })"
          >
            <el-icon><CollectionTag /></el-icon>
          </button>
        </el-tooltip>
        <div class="bottom-divider"></div>
        <div class="theme-switch theme-switch-in-bar" role="group" aria-label="Main theme mode">
          <el-tooltip content="亮色模式" placement="top" :show-after="100" :hide-after="0">
            <button
              class="bottom-btn theme-switch-btn"
              :class="{ active: themeMode === 'light' }"
              aria-label="切换亮色模式"
              @click="setThemeMode('light')"
            >
              <el-icon><Sunny /></el-icon>
            </button>
          </el-tooltip>
          <el-tooltip content="暗色模式" placement="top" :show-after="100" :hide-after="0">
            <button
              class="bottom-btn theme-switch-btn"
              :class="{ active: themeMode === 'dark' }"
              aria-label="切换暗色模式"
              @click="setThemeMode('dark')"
            >
              <el-icon><MoonNight /></el-icon>
            </button>
          </el-tooltip>
        </div>
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
  TrendCharts,
  Star,
  CollectionTag,
  Sunny,
  MoonNight
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
    TrendCharts,
    Star,
    CollectionTag,
    Sunny,
    MoonNight
  },
  setup() {
    const router = useRouter()
    const userInfo = ref(null)
    const themeMode = ref(localStorage.getItem('app-theme-mode') === 'dark' ? 'dark' : 'light')

    const applyThemeToDom = (mode) => {
      document.documentElement.setAttribute('data-app-theme', mode)
      document.body.classList.remove('app-theme-light', 'app-theme-dark')
      document.body.classList.add(mode === 'dark' ? 'app-theme-dark' : 'app-theme-light')
    }

    const setThemeMode = (mode) => {
      if (mode !== 'light' && mode !== 'dark') return
      themeMode.value = mode
      localStorage.setItem('app-theme-mode', mode)
      applyThemeToDom(mode)
    }

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
      applyThemeToDom(themeMode.value)
    })

    return {
      userInfo,
      themeMode,
      setThemeMode,
      handleUserCommand,
      hasAdminOrLibrarianRole,
      isAdmin
    }
  }
}
</script>

<style scoped>
.layout-container {
  --layout-bg: #f5f5f5;
  --panel-bg: rgba(255, 255, 255, 0.95);
  --panel-border: rgba(15, 23, 42, 0.08);
  --panel-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  --btn-fg: #6b7280;
  --btn-hover-fg: #111827;
  --btn-hover-bg: rgba(0, 0, 0, 0.05);
  --btn-active-bg: rgba(59, 130, 246, 0.1);
  --btn-active-fg: #3b82f6;
  --divider-color: #e5e7eb;
  --user-fg: #374151;
  height: 100vh;
}

.main {
  background-color: var(--layout-bg);
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
  background: var(--panel-bg);
  backdrop-filter: blur(12px);
  border-radius: 9999px;
  padding: 12px 20px;
  box-shadow: var(--panel-shadow);
  border: 1px solid var(--panel-border);
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
  color: var(--btn-fg);
  font-size: 20px;
  padding: 0 12px;
  border-radius: 20px;
  transition: all 0.2s;
}

.bottom-btn:hover {
  background: var(--btn-hover-bg);
  color: var(--btn-hover-fg);
}

.bottom-btn.active {
  background: var(--btn-active-bg);
  color: var(--btn-active-fg);
}

.bottom-divider {
  width: 1px;
  height: 24px;
  background: var(--divider-color);
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
  color: var(--user-fg);
  white-space: nowrap;
}

.theme-switch {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0;
  border-radius: 999px;
}

.theme-switch-in-bar .theme-switch-btn {
  min-width: 40px;
  height: 40px;
  padding: 0;
}

.theme-switch-btn:hover {
  background: var(--btn-hover-bg);
  color: var(--btn-hover-fg);
}

.theme-switch-btn.active {
  background: var(--btn-active-bg);
  color: var(--btn-active-fg);
}

.layout-container.theme-dark {
  --layout-bg: #0f172a;
  --panel-bg: rgba(15, 23, 42, 0.88);
  --panel-border: rgba(148, 163, 184, 0.24);
  --panel-shadow: 0 10px 30px rgba(2, 6, 23, 0.55);
  --btn-fg: #cbd5e1;
  --btn-hover-fg: #f8fafc;
  --btn-hover-bg: rgba(148, 163, 184, 0.18);
  --btn-active-bg: rgba(59, 130, 246, 0.26);
  --btn-active-fg: #93c5fd;
  --divider-color: rgba(148, 163, 184, 0.35);
  --user-fg: #e2e8f0;
}

.layout-container.theme-dark :deep(.book-gallery-page) {
  background: linear-gradient(to bottom right, #0f172a, #1e293b);
}

/* 自定义tooltip样式 */
:deep(.el-tooltip__popper) {
  font-size: 16px !important;
  padding: 10px 16px !important;
  font-weight: 500 !important;
}
</style>
