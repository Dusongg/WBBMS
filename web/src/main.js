import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'
import axios from 'axios'

// 配置axios
axios.defaults.baseURL = 'http://localhost:8888/api'
axios.defaults.timeout = 10000

// 请求拦截器
axios.interceptors.request.use(
  config => {
    // 添加token到请求头
    const token = localStorage.getItem('bookadmin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
axios.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    console.error('请求错误:', error)
    // 如果是401错误，清除token并跳转到登录页
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('bookadmin_token')
      localStorage.removeItem('bookadmin_user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

const app = createApp(App)
app.use(ElementPlus)
app.use(router)
app.config.globalProperties.$http = axios
app.mount('#app')

