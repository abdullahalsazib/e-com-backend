package middlewares

import (
	"net/http"
	"strings"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/abdullahalsazib/e-com-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var jwtSecret = []byte("your_secret_key") // set your secret key
func AuthMiddleware(DB *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Fetch user with roles (IMPORTANT: Preload)
		var user models.User
		if err := DB.Preload("Roles").First(&user, claims.UserID).Error; err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "User not found"})
			return
		}

		// Extract role slugs
		var roleSlugs []string
		for _, role := range user.Roles {
			roleSlugs = append(roleSlugs, role.Slug)
		}

		// Save to context
		ctx.Set("currentUser", &user)
		ctx.Set("user_id", user.ID)
		ctx.Set("role", roleSlugs)

		ctx.Next()
	}
}

func AdminSellerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("role")

		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Only Admin access required 1"})
			return
		}

		roleSlice, ok := roles.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid role format"})
			return
		}

		// Check if "admin" exists
		isAdmin := false
		for _, r := range roleSlice {
			if r == "admin" {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Only Admin access required 1"})
			return
		}

		c.Next()

	}
}
