package seed

import (
	"log"

	"github.com/abdullahalsazib/e-com-backend/models"
	"gorm.io/gorm"
)

func SeedCategories(db *gorm.DB) {
	categories := []models.Category{
		{Name: "Smartphones"},
		{Name: "Laptops"},
		{Name: "Headphones"},
		{Name: "Cameras"},
	}

	for _, category := range categories {
		var existing models.Category
		// Check if already exists
		if err := db.Where("name = ?", category.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&category).Error; err != nil {
					log.Printf(" Failed to seed category %s: %v", category.Name, err)
				} else {
					log.Printf(" Category %s seeded successfully", category.Name)
				}
			}
		}
	}
}
