// 认证工具函数
export const TOKEN_KEY = 'bookadmin_token'
export const USER_INFO_KEY = 'bookadmin_user'

// 保存token
export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
}

// 获取token
export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

// 删除token
export function removeToken() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_INFO_KEY)
}

// 保存用户信息
export function setUserInfo(userInfo) {
  localStorage.setItem(USER_INFO_KEY, JSON.stringify(userInfo))
}

// 获取用户信息
export function getUserInfo() {
  const userInfo = localStorage.getItem(USER_INFO_KEY)
  return userInfo ? JSON.parse(userInfo) : null
}

// 检查是否登录
export function isLoggedIn() {
  const token = getToken()
  if (!token) return false
  
  // 解析JWT token检查是否过期
  try {
    // JWT token 格式: header.payload.signature
    const parts = token.split('.')
    if (parts.length !== 3) {
      removeToken()
      return false
    }
    
    // 解析 payload (base64 解码)
    const payload = JSON.parse(atob(parts[1]))
    
    // 检查是否过期 (exp 是秒级时间戳)
    if (payload.exp && payload.exp * 1000 < Date.now()) {
      // token 已过期，清除 token
      removeToken()
      return false
    }
    
    return true
  } catch (error) {
    console.error('解析token失败:', error)
    removeToken()
    return false
  }
}

// 检查角色权限
export function hasRole(role) {
  const userInfo = getUserInfo()
  if (!userInfo) return false
  return userInfo.role === role
}

// 检查是否有管理员或图书管理员权限
export function hasAdminOrLibrarianRole() {
  const userInfo = getUserInfo()
  if (!userInfo) return false
  return userInfo.role === 'admin' || userInfo.role === 'librarian'
}

// 检查是否是管理员
export function isAdmin() {
  const userInfo = getUserInfo()
  if (!userInfo) return false
  return userInfo.role === 'admin'
}

