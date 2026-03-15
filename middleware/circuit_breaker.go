package middleware

import (
	"bookadmin/global"
	"bookadmin/resilience"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// skipBreakerPaths 不进行熔断检查的路径
var skipBreakerPaths = map[string]bool{
	"/healthz":     true,
	"/readyz":      true,
	"/metrics":     true,
	"/api/healthz": true,
	"/api/readyz":  true,
}

// CircuitBreaker 熔断器中间件：当 DB/Redis 熔断打开时，快速返回 503
func CircuitBreaker() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.GVA_CONFIG == nil || !global.GVA_CONFIG.Resilience.CircuitBreaker.Enabled {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if skipBreakerPaths[path] {
			c.Next()
			return
		}
		if resilience.DBState() == gobreaker.StateOpen || resilience.RedisState() == gobreaker.StateOpen {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code": 503,
				"msg":  "服务暂时不可用，请稍后重试",
			})
			return
		}
		c.Next()
	}
}
