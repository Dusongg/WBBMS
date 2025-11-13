import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { isLoggedIn, getUserInfo } from '../utils/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    redirect: '/books',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'books',
        name: 'BookList',
        component: () => import('../views/BookList.vue'),
        meta: { title: '图书管理' }
      },
      {
        path: 'ranking',
        name: 'RankingList',
        component: () => import('../views/RankingList.vue'),
        meta: { title: '图书榜单' }
      },
      {
        path: 'readers',
        name: 'ReaderList',
        component: () => import('../views/ReaderList.vue'),
        meta: { title: '读者管理', requiresRole: ['admin', 'librarian'] }
      },
      {
        path: 'borrow',
        name: 'BorrowList',
        component: () => import('../views/BorrowList.vue'),
        meta: { title: '借还管理', requiresRole: ['admin', 'librarian'] }
      },
      {
        path: 'my-borrow',
        name: 'MyBorrowList',
        component: () => import('../views/MyBorrowList.vue'),
        meta: { title: '我的借阅' }
      },
      {
        path: 'statistics',
        name: 'Statistics',
        component: () => import('../views/Statistics.vue'),
        meta: { title: '统计查询', requiresRole: ['admin', 'librarian'] }
      },
      {
        path: 'system',
        name: 'System',
        component: () => import('../views/System.vue'),
        meta: { title: '系统管理', requiresRole: ['admin'] }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 检查是否需要认证
  if (to.meta.requiresAuth !== false) {
    if (!isLoggedIn()) {
      next('/login')
      return
    }

    // 检查角色权限
    if (to.meta.requiresRole) {
      const roles = to.meta.requiresRole
      const userInfo = getUserInfo()
      
      if (!userInfo) {
        next('/login')
        return
      }
      
      const userRole = userInfo.role
      
      // 检查用户是否有roles中的任何一个角色
      const hasPermission = roles.some(role => {
        if (role === 'admin') {
          return userRole === 'admin'
        }
        if (role === 'librarian') {
          return userRole === 'admin' || userRole === 'librarian'
        }
        return userRole === role
      })
      
      if (!hasPermission) {
        ElMessage.warning('权限不足')
        next('/')
        return
      }
    }
  }

  // 如果已登录，访问登录页则跳转到首页
  if (to.path === '/login' && isLoggedIn()) {
    next('/')
    return
  }

  next()
})

export default router
