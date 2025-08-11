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
	DB *gorm.DB
}

func NewVendorController(db *gorm.DB) *VendorController {
	return &VendorController{DB: db}
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

	userID := c.GetUint("user_id") // JWT middleware থেকে আসবে

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

	// User এবং তার Roles একসাথে লোড করবে
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
	vc.updateVendorStatus(c, "active")
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

	// Get superadmin user id from context (middleware sets it)
	actorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	actorID := actorIDVal.(uint)

	now := time.Now()
	vendor.Status = newStatus
	if newStatus == "active" {
		vendor.ApprovedBy = &actorID
		vendor.ApprovedAt = &now

		// APPROVE হলে user এর roles এ "admin" role add করা হবে
		var user models.User
		if err := vc.DB.Preload("Roles").First(&user, vendor.UserID).Error; err == nil {
			adminRole := models.Role{}
			if err := vc.DB.Where("slug = ?", "admin").First(&adminRole).Error; err == nil {
				hasAdminRole := false
				for _, role := range user.Roles {
					if role.ID == adminRole.ID {
						hasAdminRole = true
						break
					}
				}
				if !hasAdminRole {
					user.Roles = append(user.Roles, adminRole)
					if err := vc.DB.Save(&user).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign admin role to user"})
						return
					}
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin role not found"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}
	} else {
		// Reject or Suspend হলে ApproveBy/ApproveAt NULL set করবে
		vendor.ApprovedBy = nil
		vendor.ApprovedAt = nil
		// (Optional) তুমি চাইলে reject/suspend হলে "admin" role remove করতে পারো,
		// তবে সাধারনত remove করা হয় না, কারণ user অন্য vendor/admin হিসাবেও থাকতে পারে
	}

	if err := vc.DB.Save(&vendor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vendor status"})
		return
	}

	// Create audit log entry
	oldValJSON, _ := json.Marshal(map[string]string{"status": oldStatus})
	newValJSON, _ := json.Marshal(map[string]string{"status": newStatus})

	auditLog := models.AuditLog{
		ActorID:  &actorID,
		Action:   "update_vendor_status",
		Resource: "vendor:" + idStr,
		OldValue: string(oldValJSON),
		NewValue: string(newValJSON),
		// CreatedAt: now, // GORM নিজে handle করে
	}

	if err := vc.DB.Create(&auditLog).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Vendor status updated but audit log failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vendor status updated successfully"})
}
