package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SuperAdminMiddleware(DB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Only Admin access required 1"})
			return
		}

		var isSuperAdmin bool

		switch v := roleValue.(type) {
		case string:
			isSuperAdmin = v == "superadmin"
		case []string:
			for _, r := range v {
				if r == "superadmin" {
					isSuperAdmin = true
					break
				}
			}
		case []interface{}:
			for _, r := range v {
				if str, ok := r.(string); ok && str == "superadmin" {
					isSuperAdmin = true
					break
				}
			}
		default:
			isSuperAdmin = false
		}

		if !isSuperAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Only Admin access required 1"})
			return
		}

		c.Next()
	}
}
