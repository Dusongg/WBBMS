package middleware

import (
	"bookadmin/global"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// skipPaths 不限流的路径（健康检查、就绪探针、指标）
var skipPaths = map[string]bool{
	"/healthz":     true,
	"/readyz":      true,
	"/metrics":     true,
	"/api/healthz": true,
	"/api/readyz":  true,
}

// RateLimit 全局限流中间件（Token Bucket）
func RateLimit() gin.HandlerFunc {
	if global.GVA_CONFIG == nil || !global.GVA_CONFIG.Resilience.RateLimit.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	cfg := global.GVA_CONFIG.Resilience.RateLimit
	rps := rate.Limit(cfg.RPS)
	if rps <= 0 {
		rps = 100
	}
	burst := cfg.Burst
	if burst <= 0 {
		burst = 50
	}

	var (
		globalLimiter = rate.NewLimiter(rps, burst)
		perIPLimiters = make(map[string]*rate.Limiter)
		mu            sync.RWMutex
	)

	getLimiter := func(key string) *rate.Limiter {
		mu.RLock()
		lim, ok := perIPLimiters[key]
		mu.RUnlock()
		if ok {
			return lim
		}
		mu.Lock()
		defer mu.Unlock()
		if lim, ok = perIPLimiters[key]; ok {
			return lim
		}
		lim = rate.NewLimiter(rps, burst)
		perIPLimiters[key] = lim
		return lim
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if skipPaths[path] {
			c.Next()
			return
		}

		var limiter *rate.Limiter
		if cfg.PerIP {
			limiter = getLimiter(c.ClientIP())
		} else {
			limiter = globalLimiter
		}

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

