package middleware

import (
	"go-cookbook/internal/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware(jwtService jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Authorization 字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "error", "error": "未授权"})
			c.Abort()
			return
		}
		// 提取 JWT 令牌
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "error", "error": "未授权"})
			c.Abort()
			return
		}
		// 解析 JWT 令牌
		claims, err := jwtService.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "error", "error": "未授权"})
			c.Abort()
			return
		}
		// 检查角色是否匹配
		if claims.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "error", "error": "权限不足"})
			c.Abort()
			return
		}
		// 继续处理请求
		c.Next()
	}
}
