import axios from 'axios'

/**
 * 切换点赞状态
 * @param {number} bookId - 图书ID
 */
export function toggleLike(bookId) {
  return axios.post(`/like/toggle/${bookId}`)
}

/**
 * 查询点赞状态
 * @param {number} bookId - 图书ID
 */
export function getLikeStatus(bookId) {
  return axios.get(`/like/status/${bookId}`)
}

/**
 * 批量查询点赞状态
 * @param {Array<number>} bookIds - 图书ID列表
 */
export function batchGetLikeStatus(bookIds) {
  return axios.get('/like/batch-status', {
    params: {
      bookIds: bookIds.join(',')
    }
  })
}

/**
 * 获取用户点赞列表
 * @param {number} page - 页码
 * @param {number} pageSize - 每页数量
 */
export function getUserLikeList(page = 1, pageSize = 10) {
  return axios.get('/like/list', {
    params: { page, pageSize }
  })
}

