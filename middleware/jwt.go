package middleware

import (
	"bookadmin/model"
	"bookadmin/model/common/response"
	"bookadmin/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(200, response.FailWithMessage("未登录或非法访问"))
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(200, response.FailWithMessage("未登录或非法访问"))
			c.Abort()
			return
		}

		// 解析token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(200, response.FailWithMessage("未登录或非法访问"))
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole 要求特定角色
func RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(200, response.FailWithMessage("未登录或非法访问"))
			c.Abort()
			return
		}

		userRole, ok := role.(model.UserRole)
		if !ok {
			c.JSON(200, response.FailWithMessage("未登录或非法访问"))
			c.Abort()
			return
		}

		// 检查是否有权限
		hasPermission := false
		for _, r := range roles {
			if userRole == r {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(200, response.FailWithMessage("权限不足"))
			c.Abort()
			return
		}

		c.Next()
	}
}
