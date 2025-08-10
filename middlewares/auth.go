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

		// Fetch user by ID from claims
		var user models.User
		if err := DB.First(&user, claims.UserID).Error; err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "User not found"})
			return
		}

		// Extract role from user roles
		var roleSlugs []string
		for _, role := range user.Roles {
			roleSlugs = append(roleSlugs, role.Slug)
		}

		// Store user info in context
		ctx.Set("currentUser", &user)
		ctx.Set("user_id", user.ID)
		ctx.Set("role", roleSlugs)

		ctx.Next()
	}
}

func AdminSellerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Only Admin access required"})
			return
		}

		roleStr, ok := role.(string)
		if !ok || roleStr != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Only Admin access required"})
			return
		}
		c.Next()
	}
}

// func AdminMiddleware() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		role := ctx.GetString("userRole")
// 		if role != "admin" && role != "superadmin" {
// 			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
// 			return
// 		}
// 		ctx.Next()
// 	}
// }
//
// func AdminOnly() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
// 			c.Abort()
// 			return
// 		}
//
// 		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
// 		if tokenString == authHeader {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token missing"})
// 			c.Abort()
// 			return
// 		}
//
// 		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
// 			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, jwt.ErrSignatureInvalid
// 			}
// 			return jwtSecret, nil
// 		})
//
// 		if err != nil || !token.Valid {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 			c.Abort()
// 			return
// 		}
//
// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok || claims["role"] == nil {
// 			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token claims"})
// 			c.Abort()
// 			return
// 		}
//
// 		role := claims["role"].(string)
// 		if role != "admin" && role != "superAdmin" {
// 			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: admin only"})
// 			c.Abort()
// 			return
// 		}
//
// 		// Store role/user_id in context if needed
// 		c.Set("role", role)
// 		c.Next()
// 	}
// }
