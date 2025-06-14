package helper

import (
	"github.com/abdullahalsazib/e-com-backend/models"
	"gorm.io/gorm"
)

type ProductWishlistHelper struct {
	DB *gorm.DB
}

func NewProductWishlistHelper(DB *gorm.DB) ProductWishlistHelper {
	return ProductWishlistHelper{DB}
}

// productExists checks if a product with the given ID exists
func (ph *ProductWishlistHelper) ProductExists(productID uint) bool {
	// Implement based on your product service
	// This could be a database check or API call
	var count int64
	ph.DB.Model(&models.Product{}).Where("id = ?", productID).Count(&count)
	return count > 0
}

// getProductDetails fetches product information
func (ph *ProductWishlistHelper) GetProductDetails(productID uint) (map[string]interface{}, error) {
	// Implement based on your product service
	var product models.Product
	if err := ph.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":    product.ID,
		"name":  product.Name,
		"price": product.Price,
		// Add other product fields as needed
	}, nil
}
