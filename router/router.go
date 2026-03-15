package router

import (
	"bookadmin/global"
	"bookadmin/middleware"
	"bookadmin/observability"
	"bookadmin/resilience"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

// traceIDHeaderMiddleware 将 trace_id 写入响应头，便于前端/日志关联（在 c.Next 前写入，确保 header 生效）
func traceIDHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().HasTraceID() {
			c.Header("X-Trace-ID", span.SpanContext().TraceID().String())
		}
		c.Next()
	}
}

var instanceID = initInstanceID()

// readyzHandler 就绪探针：DB/Redis Ping 驱动熔断状态；withMetrics 时写 DB/Redis 指标
func readyzHandler(withMetrics bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ready := global.GVA_DB != nil
		if global.GVA_DB != nil {
			if sqlDB, err := global.GVA_DB.DB(); err == nil {
				_, pingErr := resilience.DBExecute(func() (interface{}, error) {
					ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
					defer cancel()
					return nil, sqlDB.PingContext(ctx)
				})
				if pingErr != nil {
					ready = false
				}
				if withMetrics {
					stats := sqlDB.Stats()
					observability.SetDBStats(stats.OpenConnections, stats.Idle, stats.InUse)
				}
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
			_, pingErr := resilience.RedisExecute(func() (interface{}, error) {
				return nil, global.GVA_REDIS.Ping(ctx).Err()
			})
			if pingErr != nil {
				ready = false
			} else if withMetrics {
				observability.SetRedisPing(time.Since(start))
			}
		}
		statusCode := http.StatusOK
		status := "ready"
		if !ready {
			statusCode = http.StatusServiceUnavailable
			status = "not_ready"
		}
		c.JSON(statusCode, gin.H{"status": status})
	}
}

func initInstanceID() string {
	if id := os.Getenv("APP_INSTANCE_ID"); id != "" {
		return id
	}
	hostname, _ := os.Hostname()
	if hostname != "" {
		return hostname
	}
	return "default"
}

func InitRouter() *gin.Engine {
	if global.GVA_CONFIG != nil && global.GVA_CONFIG.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	Router := gin.New()
	Router.Use(gin.Recovery())
	Router.Use(func(c *gin.Context) {
		c.Header("X-Instance-ID", instanceID)
		if observability.TracerEnabled() {
			c.Header("X-Tracing-Status", "enabled")
		} else {
			c.Header("X-Tracing-Status", "disabled")
		}
		c.Next()
	})
	if observability.TracerEnabled() {
		Router.Use(otelgin.Middleware("bookadmin-api"))
		Router.Use(traceIDHeaderMiddleware())
	}
	Router.Use(middleware.RequestContext())
	if global.GVA_CONFIG != nil && global.GVA_CONFIG.Resilience.RateLimit.Enabled {
		Router.Use(middleware.RateLimit())
	}
	if global.GVA_CONFIG != nil && global.GVA_CONFIG.Resilience.CircuitBreaker.Enabled {
		Router.Use(middleware.CircuitBreaker())
	}
	Router.Use(middleware.CORS())

	Router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"instance_id": instanceID,
			"time":        time.Now().Format(time.RFC3339),
		})
	})

	Router.GET("/readyz", readyzHandler(true))

	if global.GVA_CONFIG == nil || global.GVA_CONFIG.Metrics.Enabled {
		Router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// API路由组
	apiRouter := Router.Group("api")
	{
		// 健康检查（供验证脚本 /api/healthz 使用，nginx 会代理此路径）
		apiRouter.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":          "ok",
				"instance_id":     instanceID,
				"time":            time.Now().Format(time.RFC3339),
				"tracing_enabled": observability.TracerEnabled(),
			})
		})

		// 就绪探针（供熔断验证等使用，会执行 DB/Redis Ping 驱动熔断状态）
		apiRouter.GET("/readyz", readyzHandler(false))

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
