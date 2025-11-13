import axios from 'axios'

/**
 * 切换收藏状态
 * @param {number} bookId - 图书ID
 */
export function toggleFavorite(bookId) {
  return axios.post(`/favorite/toggle/${bookId}`)
}

/**
 * 查询收藏状态
 * @param {number} bookId - 图书ID
 */
export function getFavoriteStatus(bookId) {
  return axios.get(`/favorite/status/${bookId}`)
}

/**
 * 批量查询收藏状态
 * @param {Array<number>} bookIds - 图书ID列表
 */
export function batchGetFavoriteStatus(bookIds) {
  return axios.get('/favorite/batch-status', {
    params: {
      bookIds: bookIds.join(',')
    }
  })
}

/**
 * 获取用户收藏列表
 * @param {number} page - 页码
 * @param {number} pageSize - 每页数量
 */
export function getUserFavoriteList(page = 1, pageSize = 10) {
  return axios.get('/favorite/list', {
    params: { page, pageSize }
  })
}

