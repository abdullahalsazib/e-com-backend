package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/abdullahalsazib/e-com-backend/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(DB *gorm.DB) AuthController {
	return AuthController{DB}
}

func (ac *AuthController) Register(c *gin.Context) {
	var payload struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=4"` // min >= 8
		Role     string `json:"role" binding:"required,oneof=admin user superadmin"`
	}

	// Bind -> json and validate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check for existing user
	var existingUser models.User
	if err := ac.DB.Where("email = ?", payload.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User with this email already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error cheching for existing user", "details": err.Error()})
		return
	}

	// Hash password

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Faild to hash password", "details": err.Error()})
		return
	}

	// create user
	newUser := models.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Role:     payload.Role,
		Password: string(hashedPassword),
	}
	if err := ac.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Faild to create user", "details": err.Error()})
		return
	}
	err = utils.SendWellcomeEmail(payload.Name, payload.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "some issu for send message on email"})
	}
	// Return create user (without password)
	c.JSON(http.StatusOK, gin.H{"message": "User created succesfully"})
}

// login
func (ac *AuthController) Login(c *gin.Context) {
	var payload struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := ac.DB.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// compair the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// access_token and refresh_token
	accessToken, refreshToken, err := utils.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generate token"})
		return
	}

	// Store refresh token in database
	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour).Unix()
	token := models.Token{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshTokenExp,
		Role:      user.Role,
	}

	if err := ac.DB.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Login succesfully",
		"access_token": accessToken,
		// "refresh_token": refreshToken,
		"expires_in": 15 * 60,
		"role":       user.Role,
	})
}

// GET the Profile
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var user models.User

	if err := ac.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user profile"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get User",
		"data":    user,
	})
}

// refresh
func (ac *AuthController) Refresh(c *gin.Context) {
	var refreshData struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&refreshData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// verify Refresh( token)
	claims, err := utils.ParseToken(refreshData.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// check if Refresh token exists in database
	var token models.Token
	if err := ac.DB.Where("token = ? AND user_id = ?", refreshData.RefreshToken, claims.UserID).First(&token).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Check if refresh token is expired
	if time.Now().Unix() > token.ExpiresAt {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	// Generate new token pair
	accessToken, newRefreshToken, err := utils.GenerateTokenPair(claims.UserID, claims.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
		return
	}

	// Delete old refresh token
	ac.DB.Delete(&token)

	// Store new refresh token
	newRefreshTokenExp := time.Now().Add(7 * 24 * time.Hour).Unix()
	newToken := models.Token{
		UserID:    claims.UserID,
		Token:     newRefreshToken,
		ExpiresAt: newRefreshTokenExp,
	}

	if err := ac.DB.Create(&newToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    15 * 60, // 15 minutes
	})
}

func (ac *AuthController) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Delete all refresh tokens for this user
	if err := ac.DB.Where("user_id = ?", userID).Delete(&models.Token{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error logging out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
