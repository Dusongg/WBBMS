package model

// RankingType 榜单类型
type RankingType string

const (
	RankingTypeLike     RankingType = "like"     // 点赞榜
	RankingTypeFavorite RankingType = "favorite" // 收藏榜
)

// RankingPeriod 榜单周期
type RankingPeriod string

const (
	RankingPeriodWeek  RankingPeriod = "week"  // 周榜
	RankingPeriodMonth RankingPeriod = "month" // 月榜
)

// RankingItem 榜单项
type RankingItem struct {
	Rank  int   `json:"rank"`  // 排名
	BookID uint  `json:"book_id"` // 图书ID
	Book  *Book `json:"book,omitempty"` // 图书详情
	Score int64 `json:"score"` // 分数（点赞数或收藏数）
}

// RankingResponse 榜单响应
type RankingResponse struct {
	Type      RankingType    `json:"type"`      // 榜单类型
	Period    RankingPeriod  `json:"period"`    // 榜单周期
	PeriodKey string         `json:"period_key"` // 周期标识（如：2025-W45）
	Items     []RankingItem  `json:"items"`     // 榜单数据
	Total     int            `json:"total"`     // 总数
	UpdatedAt string         `json:"updated_at"` // 更新时间
}

// RankingRequest 榜单查询请求
type RankingRequest struct {
	Type   RankingType   `json:"type" form:"type" binding:"required,oneof=like favorite"` // 榜单类型
	Period RankingPeriod `json:"period" form:"period" binding:"required,oneof=week month"` // 榜单周期
	Limit  int           `json:"limit" form:"limit"` // 返回数量，默认100
}

