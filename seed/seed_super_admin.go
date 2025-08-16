package seed

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/abdullahalsazib/e-com-backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedSuperAdmin(db *gorm.DB) {

	var existingRole models.Role
	if err := db.Where("slug = ?", "superadmin").First(&existingRole).Error; err == nil {
		// Role already exists, skip insertion or update if needed
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Role doesn't exist, so create it
		newRole := models.Role{Name: "Super Admin", Slug: "superadmin"}
		if err := db.Create(&newRole).Error; err != nil {
			log.Fatalf("Failed to create superadmin role: %v", err)
		}
	}

	role := models.Role{}
	if err := db.Where("slug = ?", "superadmin").First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			role = models.Role{
				Name: "Super Admin",
				Slug: "superadmin",
			}
			if err := db.Create(&role).Error; err != nil {
				log.Fatalf("Failed to create superadmin role: %v", err)
			}
			log.Println("Super Admin role created")
		} else {
			log.Fatalf("Role query error: %v", err)
		}
	} else {
		log.Println("Super Admin role already exists")
	}

	var user models.User
	email := os.Getenv("SUPER_USER_EMAIL")
	password := os.Getenv("SUPER_USER_PASSWORD")
	if email == "" {
		log.Fatal("SUPER_USER_EMAIL is not found")
	}
	if password == "" {
		log.Fatal("SUPER_USER_PASSWORD is not found")
	}

	err := db.Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		user = models.User{
			Name:     "Super Admin",
			Email:    email,
			Password: string(hashedPassword),
			IsActive: true,
			Model: gorm.Model{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		if err := db.Create(&user).Error; err != nil {
			log.Fatalf("Failed to create super admin user: %v", err)
		}
		log.Println("Super Admin user created")
	} else if err != nil {
		log.Fatalf("User query error: %v", err)
	} else {
		log.Println("Super Admin user already exists")
	}

	var userRole models.UserRole
	err = db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole).Error
	if err == gorm.ErrRecordNotFound {
		userRole = models.UserRole{
			UserID: user.ID,
			RoleID: role.ID,
		}
		if err := db.Create(&userRole).Error; err != nil {
			log.Fatalf("Failed to assign super admin role: %v", err)
		}
		log.Println("Super Admin role assigned to user")
	} else if err != nil {
		log.Fatalf("UserRole query error: %v", err)
	} else {
		log.Println("User already has Super Admin role")
	}
}
