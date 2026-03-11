package router

import (
	"bookadmin/global"
	"bookadmin/middleware"
	"bookadmin/observability"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitRouter() *gin.Engine {
	if global.GVA_CONFIG != nil && global.GVA_CONFIG.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	Router := gin.New()
	Router.Use(gin.Recovery())
	Router.Use(middleware.RequestContext())
	Router.Use(middleware.CORS())

	Router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	Router.GET("/readyz", func(c *gin.Context) {
		ready := global.GVA_DB != nil
		if global.GVA_DB != nil {
			if sqlDB, err := global.GVA_DB.DB(); err == nil {
				if err := sqlDB.PingContext(c.Request.Context()); err != nil {
					ready = false
				}
				stats := sqlDB.Stats()
				observability.SetDBStats(stats.OpenConnections, stats.Idle, stats.InUse)
			} else {
				ready = false
			}
		}

		if global.GVA_CONFIG != nil && global.GVA_CONFIG.Redis.Enabled && global.GVA_REDIS == nil {
			ready = false
		}

		if ready && global.GVA_REDIS != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()

			start := time.Now()
			if err := global.GVA_REDIS.Ping(ctx).Err(); err != nil {
				ready = false
			} else {
				observability.SetRedisPing(time.Since(start))
			}
		}

		statusCode := http.StatusOK
		status := "ready"
		if !ready {
			statusCode = http.StatusServiceUnavailable
			status = "not_ready"
		}

		c.JSON(statusCode, gin.H{
			"status": status,
		})
	})

	if global.GVA_CONFIG == nil || global.GVA_CONFIG.Metrics.Enabled {
		Router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// API路由组
	apiRouter := Router.Group("api")
	{
		// 认证相关（无需JWT）
		InitAuthRouter(apiRouter)

		// 图书管理（需要JWT，部分接口需要管理员权限）
		InitBookRouter(apiRouter)

		// 分类管理
		InitCategoryRouter(apiRouter)

		// 读者管理
		InitReaderRouter(apiRouter)

		// 借还管理
		InitBorrowRouter(apiRouter)

		// 统计查询
		InitStatisticsRouter(apiRouter)

		// 系统管理
		InitSystemRouter(apiRouter)

		// 点赞功能
		InitLikeRouter(apiRouter)

		// 收藏功能
		InitFavoriteRouter(apiRouter)

		// 榜单功能
		InitRankingRouter(apiRouter)

		// 预约管理
		InitReservationRouter(apiRouter)

		// 罚款管理
		InitFineRouter(apiRouter)

		// 黑名单管理
		InitBlacklistRouter(apiRouter)

		// 消息管理
		InitMessageRouter(apiRouter)
	}

	return Router
}
