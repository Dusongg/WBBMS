import axios from 'axios'

/**
 * 获取榜单数据
 * @param {string} type - 榜单类型：'like' 点赞榜, 'favorite' 收藏榜
 * @param {string} period - 榜单周期：'week' 周榜, 'month' 月榜
 * @param {number} limit - 返回数量，默认100
 */
export function getRanking(type, period, limit = 100) {
  return axios.get('/ranking/list', {
    params: { type, period, limit }
  })
}

/**
 * 重建榜单（管理员功能）
 * @param {string} type - 榜单类型：'like' 点赞榜, 'favorite' 收藏榜
 * @param {string} period - 榜单周期：'week' 周榜, 'month' 月榜
 */
export function rebuildRanking(type, period) {
  return axios.post('/ranking/rebuild', { type, period })
}
