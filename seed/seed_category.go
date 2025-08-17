package seed

import (
	"log"
	"strings"

	"github.com/abdullahalsazib/e-com-backend/models"
	"gorm.io/gorm"
)

// Helper function to generate slug from name
func generateSlug(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

func SeedCategories(db *gorm.DB) {
	categories := []models.Category{
		{Name: "Mobile", Slug: generateSlug("Mobile"), Description: "Smartphones and mobile devices"},
		{Name: "Laptop", Slug: generateSlug("Laptop"), Description: "All types of laptops"},
		{Name: "Accessories", Slug: generateSlug("Accessories"), Description: "Phone and laptop accessories"},
		{Name: "Home Appliances", Slug: generateSlug("Home Appliances"), Description: "Appliances for home use"},
	}

	for _, category := range categories {
		var existing models.Category
		// Check if category already exists
		if err := db.Where("name = ?", category.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&category).Error; err != nil {
					log.Printf("Failed to seed category %s: %v", category.Name, err)
				} else {
					log.Printf("Category %s seeded successfully", category.Name)
				}
			} else {
				log.Printf("Error checking category %s: %v", category.Name, err)
			}
		} else {
			log.Printf("Category %s already exists, skipping", category.Name)
		}
	}
}
