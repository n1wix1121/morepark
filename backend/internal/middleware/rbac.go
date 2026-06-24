package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRoles проверяет, что у пользователя есть одна из указанных ролей
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userRole := role.(string)

		// Проверяем, есть ли роль в списке разрешённых
		for _, r := range roles {
			if r == userRole {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":          "У вас нет доступа к этому ресурсу",
			"required_roles": roles,
			"your_role":      userRole,
		})
	}
}
