package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VendorController struct {
	DB       *gorm.DB
	AuthCtrl *AuthController
}

func NewVendorController(db *gorm.DB, authCtrl *AuthController) *VendorController {
	return &VendorController{
		DB:       db,
		AuthCtrl: authCtrl,
	}
}

// User applies to become a vendor
func (vc *VendorController) VendorApply(c *gin.Context) {
	var input struct {
		ShopName string `json:"shop_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id") // Comes from JWT middleware

	// Check if vendor already applied
	var existing models.Vendor
	if err := vc.DB.Where("user_id = ?", userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already applied as vendor"})
		return
	}

	vendor := models.Vendor{
		UserID:   userID,
		ShopName: input.ShopName,
		Status:   "pending",
	}

	if err := vc.DB.Create(&vendor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not submit vendor application"})
		return
	}
	if err := vc.DB.Preload("User").First(&vendor, vendor.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preload user"})
		return
	}

	// Audit log
	audit := models.AuditLog{
		ActorID:  &userID,
		Action:   "vendor_apply",
		Resource: "vendor",
		OldValue: "",
		NewValue: input.ShopName,
	}
	vc.DB.Create(&audit)

	c.JSON(http.StatusOK, gin.H{"message": "Vendor application submitted", "vendor": vendor})
}

// List vendors (with optional status filter)
func (vc *VendorController) ListVendors(c *gin.Context) {
	status := c.Query("status")
	var vendors []models.Vendor
	query := vc.DB

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Load User and their Roles together
	query = query.Preload("User.Roles")

	err := query.Find(&vendors).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vendors"})
		return
	}
	

	c.JSON(http.StatusOK, vendors)
}

// Get single vendor details by id
func (vc *VendorController) GetVendor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vendor ID"})
		return
	}

	var vendor models.Vendor
	if err := vc.DB.Preload("User").First(&vendor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
		return
	}

	c.JSON(http.StatusOK, vendor)
}

// Approve vendor (superadmin only)
func (vc *VendorController) ApproveVendor(c *gin.Context) {
	vc.updateVendorStatus(c, "approved")
}

// Reject vendor (superadmin only)
func (vc *VendorController) RejectVendor(c *gin.Context) {
	vc.updateVendorStatus(c, "rejected")
}

// Suspend vendor (superadmin only)
func (vc *VendorController) SuspendVendor(c *gin.Context) {
	vc.updateVendorStatus(c, "suspended")
}

func (vc *VendorController) updateVendorStatus(c *gin.Context, newStatus string) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vendor ID"})
		return
	}

	var vendor models.Vendor
	if err := vc.DB.First(&vendor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
		return
	}

	oldStatus := vendor.Status
	if oldStatus == newStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vendor already has status " + newStatus})
		return
	}

	actorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	actorID := actorIDVal.(uint)

	now := time.Now()
	vendor.Status = newStatus

	var user models.User
	if err := vc.DB.Preload("Roles").First(&user, vendor.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	if newStatus == "approved" {
		vendor.ApprovedBy = &actorID
		vendor.ApprovedAt = &now

		// AuthController er helper method diye admin role add koro
		if err := vc.AuthCtrl.AddAdminRoleByUser(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add admin role"})
			return
		}

		// Ensure user role o thakbe
		if err := vc.AuthCtrl.AddRoleByUserSlug(&user, "user"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user role"})
			return
		}

	} else if newStatus == "rejected" || newStatus == "suspended" {
		vendor.ApprovedBy = nil
		vendor.ApprovedAt = nil

		// Admin role remove koro
		if err := vc.AuthCtrl.RemoveAdminRoleByUser(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove admin role"})
			return
		}

		// Ensure user role thakbe
		if err := vc.AuthCtrl.AddRoleByUserSlug(&user, "user"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure user role"})
			return
		}
	} else {
		vendor.ApprovedBy = nil
		vendor.ApprovedAt = nil
	}

	if err := vc.DB.Save(&vendor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vendor status"})
		return
	}

	oldValJSON, _ := json.Marshal(map[string]string{"status": oldStatus})
	newValJSON, _ := json.Marshal(map[string]string{"status": newStatus})

	auditLog := models.AuditLog{
		ActorID:  &actorID,
		Action:   "update_vendor_status",
		Resource: "vendor:" + idStr,
		OldValue: string(oldValJSON),
		NewValue: string(newValJSON),
	}

	if err := vc.DB.Create(&auditLog).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Vendor status updated but audit log failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vendor status updated successfully"})
}
