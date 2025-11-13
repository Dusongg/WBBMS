package constants

import "fmt"

// ============================================
// Redis Key 命名规范
// ============================================

// 用户点赞状态 (Set)
// key: user:likes:{user_id}
// value: Set{book_id1, book_id2, ...}
// 过期时间: 24小时
func KeyUserLikes(userID uint) string {
	return fmt.Sprintf("user:likes:%d", userID)
}

// 用户收藏状态 (Set)
// key: user:favorites:{user_id}
// value: Set{book_id1, book_id2, ...}
// 过期时间: 24小时
func KeyUserFavorites(userID uint) string {
	return fmt.Sprintf("user:favorites:%d", userID)
}

// 图书统计信息 (Hash)
// key: book:stats:{book_id}
// fields: {like_count: 数量, favorite_count: 数量}
// 过期时间: 1小时
func KeyBookStats(bookID uint) string {
	return fmt.Sprintf("book:stats:%d", bookID)
}

// 点赞周榜 (Sorted Set - ZSet)
// key: rank:likes:week:{年}-W{周数}
// score: 点赞数
// member: book_id
// 过期时间: 8天
func KeyRankLikesWeek(period string) string {
	return fmt.Sprintf("rank:likes:week:%s", period)
}

// 点赞月榜 (Sorted Set - ZSet)
// key: rank:likes:month:{年}-{月}
// score: 点赞数
// member: book_id
// 过期时间: 35天
func KeyRankLikesMonth(period string) string {
	return fmt.Sprintf("rank:likes:month:%s", period)
}

// 收藏周榜 (Sorted Set - ZSet)
// key: rank:favorites:week:{年}-W{周数}
// score: 收藏数
// member: book_id
// 过期时间: 8天
func KeyRankFavoritesWeek(period string) string {
	return fmt.Sprintf("rank:favorites:week:%s", period)
}

// 收藏月榜 (Sorted Set - ZSet)
// key: rank:favorites:month:{年}-{月}
// score: 收藏数
// member: book_id
// 过期时间: 35天
func KeyRankFavoritesMonth(period string) string {
	return fmt.Sprintf("rank:favorites:month:%s", period)
}

// 点赞操作Stream (Stream)
// key: stream:like:actions
// 用于异步同步点赞操作到MySQL
const StreamLikeActions = "stream:like:actions"

// 收藏操作Stream (Stream)
// key: stream:favorite:actions
// 用于异步同步收藏操作到MySQL
const StreamFavoriteActions = "stream:favorite:actions"

// 操作锁 (String)
// key: lock:like:{user_id}:{book_id}
// 用于防止重复点赞/收藏
// 过期时间: 1秒
func KeyLikeLock(userID, bookID uint) string {
	return fmt.Sprintf("lock:like:%d:%d", userID, bookID)
}

func KeyFavoriteLock(userID, bookID uint) string {
	return fmt.Sprintf("lock:favorite:%d:%d", userID, bookID)
}

// 热点检测 (HyperLogLog)
// key: hotspot:check:{minute}
// 用于统计每分钟的操作UV
func KeyHotspotCheck(minute string) string {
	return fmt.Sprintf("hotspot:check:%s", minute)
}

// 限流 (String)
// key: rate:limit:like
// 全局限流，限制每秒点赞操作数量
const KeyRateLimitLike = "rate:limit:like"

const KeyRateLimitFavorite = "rate:limit:favorite"

// 用户级限流 (String)
// key: rate:limit:user:{user_id}:like
// 每用户每分钟最多10次点赞/收藏操作
func KeyRateLimitUserLike(userID uint) string {
	return fmt.Sprintf("rate:limit:user:%d:like", userID)
}

func KeyRateLimitUserFavorite(userID uint) string {
	return fmt.Sprintf("rate:limit:user:%d:favorite", userID)
}

// ============================================
// Redis Key 过期时间常量
// ============================================

const (
	// 用户状态缓存过期时间（24小时）
	ExpireUserStatus = 24 * 60 * 60

	// 图书统计信息过期时间（1小时）
	ExpireBookStats = 60 * 60

	// 周榜过期时间（8天）
	ExpireWeekRank = 8 * 24 * 60 * 60

	// 月榜过期时间（35天）
	ExpireMonthRank = 35 * 24 * 60 * 60

	// 操作锁过期时间（1秒）
	ExpireLock = 1

	// 热点检测过期时间（1小时）
	ExpireHotspot = 60 * 60

	// 限流窗口（1分钟）
	ExpireRateLimit = 60
)

