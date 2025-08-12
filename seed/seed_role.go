package seed

import (
	"github.com/abdullahalsazib/e-com-backend/models"
	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) {
	roles := []models.Role{
		{Name: "Admin", Slug: "admin"},
		{Name: "User", Slug: "user"},
		{Name: "Super Admin", Slug: "superadmin"},
	}

	for _, role := range roles {
		var existing models.Role
		if err := db.Where("slug = ?", role.Slug).First(&existing).Error; err == gorm.ErrRecordNotFound {
			db.Create(&role)
		}
	}
}
