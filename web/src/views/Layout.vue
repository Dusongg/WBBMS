<template>
  <el-container class="layout-container" :class="`theme-${themeMode}`">
    <el-main class="main">
      <router-view />
    </el-main>

    <!-- 底部功能栏 -->
    <div class="bottom-bar">
      <div class="bottom-bar-content">
        <!-- 导航菜单 -->
        <button
          class="bottom-link"
          :class="{ active: $route.path === '/books' && !$route.query.view }"
          @click="$router.push('/books')"
        >
          <span class="link-icon"><el-icon><Reading /></el-icon></span>
          <span class="link-title">图书管理</span>
        </button>
        <button
          class="bottom-link"
          :class="{ active: $route.path === '/ranking' }"
          @click="$router.push('/ranking')"
        >
          <span class="link-icon"><el-icon><TrendCharts /></el-icon></span>
          <span class="link-title">图书榜单</span>
        </button>
        <button
          class="bottom-link"
          :class="{ active: $route.path === '/books' && $route.query.view === 'borrow' }"
          @click="$router.push({ path: '/books', query: { view: 'borrow' } })"
        >
          <span class="link-icon"><el-icon><List /></el-icon></span>
          <span class="link-title">我的借阅</span>
        </button>
        <button
          v-if="hasAdminOrLibrarianRole()"
          class="bottom-link"
          :class="{ active: $route.path === '/readers' }"
          @click="$router.push('/readers')"
        >
          <span class="link-icon"><el-icon><User /></el-icon></span>
          <span class="link-title">读者管理</span>
        </button>
        <button
          v-if="hasAdminOrLibrarianRole()"
          class="bottom-link"
          :class="{ active: $route.path === '/borrow' }"
          @click="$router.push('/borrow')"
        >
          <span class="link-icon"><el-icon><Document /></el-icon></span>
          <span class="link-title">借还管理</span>
        </button>
        <button
          v-if="hasAdminOrLibrarianRole()"
          class="bottom-link"
          :class="{ active: $route.path === '/statistics' }"
          @click="$router.push('/statistics')"
        >
          <span class="link-icon"><el-icon><DataAnalysis /></el-icon></span>
          <span class="link-title">统计查询</span>
        </button>
        <button
          v-if="isAdmin()"
          class="bottom-link"
          :class="{ active: $route.path === '/system' }"
          @click="$router.push('/system')"
        >
          <span class="link-icon"><el-icon><Setting /></el-icon></span>
          <span class="link-title">系统管理</span>
        </button>
        
        <div class="bottom-divider"></div>
        
        <el-dropdown @command="handleUserCommand" trigger="click">
          <button class="bottom-link user-link">
            <span class="link-icon"><el-icon><UserFilled /></el-icon></span>
            <span class="link-title">{{ userInfo?.real_name || userInfo?.username || '用户' }}</span>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人信息</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <button
          class="bottom-link"
          :class="{ active: $route.path === '/books' && $route.query.view === 'like' }"
          @click="$router.push({ path: '/books', query: { view: 'like' } })"
        >
          <span class="link-icon">
            <svg class="bottom-link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" fill="currentColor" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </span>
          <span class="link-title">我喜欢的</span>
        </button>
        <button
          class="bottom-link"
          :class="{ active: $route.path === '/books' && $route.query.view === 'favorite' }"
          @click="$router.push({ path: '/books', query: { view: 'favorite' } })"
        >
          <span class="link-icon">
            <svg class="bottom-link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" fill="currentColor" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </span>
          <span class="link-title">我收藏的</span>
        </button>
        
        <div class="bottom-divider"></div>
        
        <div class="theme-switch" role="group" aria-label="Main theme mode">
          <button
            class="bottom-link theme-link"
            :class="{ active: themeMode === 'light' }"
            aria-label="切换亮色模式"
            @click="setThemeMode('light')"
          >
            <span class="link-icon"><el-icon><Sunny /></el-icon></span>
            <span class="link-title">亮色模式</span>
          </button>
          <button
            class="bottom-link theme-link"
            :class="{ active: themeMode === 'dark' }"
            aria-label="切换暗色模式"
            @click="setThemeMode('dark')"
          >
            <span class="link-icon"><el-icon><MoonNight /></el-icon></span>
            <span class="link-title">暗色模式</span>
          </button>
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

/* 底部功能栏（仿 menu / link 展开样式） */
.bottom-bar {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 100;
}

.bottom-bar-content {
  padding: 8px;
  background-color: #fff;
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  border-radius: 9999px;
  box-shadow: 0 10px 25px 0 rgba(0, 0, 0, 0.075);
  border: 1px solid var(--panel-border);
}

.bottom-link {
  display: inline-flex;
  justify-content: flex-start;
  align-items: center;
  width: 50px;
  height: 50px;
  border-radius: 9999px;
  position: relative;
  z-index: 1;
  overflow: hidden;
  transform-origin: center left;
  transition: width 0.2s ease-in;
  border: none;
  cursor: pointer;
  background: transparent;
  color: var(--btn-fg);
  font-size: 14px;
  text-decoration: none;
  padding: 0;
}

.bottom-link::before {
  position: absolute;
  z-index: -1;
  content: "";
  display: block;
  border-radius: 9999px;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
  transform: translateX(100%);
  transition: transform 0.2s ease-in;
  transform-origin: center right;
  background-color: #eee;
}

.bottom-link:hover,
.bottom-link:focus {
  outline: 0;
  width: 120px;
}

.bottom-link:hover::before,
.bottom-link:focus::before {
  transform: translateX(0);
}

.bottom-link:hover .link-title,
.bottom-link:focus .link-title {
  transform: translateX(0);
  opacity: 1;
}

.bottom-link.active::before {
  transform: translateX(0);
  background-color: rgba(59, 130, 246, 0.15);
}

.bottom-link.active {
  color: var(--btn-active-fg);
}

.link-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-left: 11px;
}

.link-icon .el-icon {
  font-size: 22px;
}

.link-icon .bottom-link-icon {
  width: 22px;
  height: 22px;
}

.link-title {
  transform: translateX(100%);
  transition: transform 0.2s ease-in;
  transform-origin: center right;
  display: block;
  text-align: left;
  margin-left: 12px;
  white-space: nowrap;
  opacity: 0;
  color: inherit;
  font-weight: 500;
}

.user-link .link-title {
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bottom-divider {
  width: 1px;
  height: 24px;
  background: var(--divider-color);
  margin: 0 4px;
  flex-shrink: 0;
}

.theme-switch {
  display: inline-flex;
  align-items: center;
  gap: 4px;
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

.layout-container.theme-dark .bottom-bar-content {
  background-color: rgba(15, 23, 42, 0.95);
  box-shadow: 0 10px 25px 0 rgba(0, 0, 0, 0.3);
}

.layout-container.theme-dark .bottom-link::before {
  background-color: rgba(148, 163, 184, 0.2);
}

.layout-container.theme-dark .bottom-link.active::before {
  background-color: rgba(59, 130, 246, 0.25);
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
