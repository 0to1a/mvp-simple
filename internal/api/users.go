package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	core "project/internal"
	"project/internal/db/sqlc"
)

// UserResponse represents a user response structure
type UserResponse struct {
	ID        int32  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	IsAdmin   bool   `json:"is_admin"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Name    string `json:"name" binding:"required"`
	IsAdmin bool   `json:"is_admin"`
}

// UserHandler handles user management operations
type UserHandler struct {
	App *core.App
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(app *core.App) *UserHandler {
	return &UserHandler{App: app}
}

// ListUsers returns all users for the company (admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Get company ID from context (set by AuthRequired middleware)
	companyID, ok := c.Get("company_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company context not found"})
		return
	}

	// Get users for the company
	users, err := h.App.Queries.ListUsers(c, companyID.(int32))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	// Transform to response format
	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			IsAdmin:   user.IsAdmin.Bool,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// CreateUser creates a new user and adds them to the company (admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Get company ID from context
	companyID, ok := c.Get("company_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company context not found"})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Check if user already exists
	existingUser, err := h.App.Queries.GetUserByEmail(c, req.Email)
	if err == nil {
		// User exists, check if they're already in the company
		checkParams := &sqlc.CheckUserInCompanyParams{
			UserID:    existingUser.ID,
			CompanyID: companyID.(int32),
		}
		inCompany, err := h.App.Queries.CheckUserInCompany(c, checkParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check user company membership"})
			return
		}
		if inCompany {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists in this company"})
			return
		}

		// Add existing user to company
		addParams := &sqlc.AddUserToCompanyParams{
			UserID:    existingUser.ID,
			CompanyID: companyID.(int32),
			IsAdmin:   sql.NullBool{Bool: req.IsAdmin, Valid: true},
		}
		err = h.App.Queries.AddUserToCompany(c, addParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add user to company"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "existing user added to company",
			"user": UserResponse{
				ID:        existingUser.ID,
				Email:     existingUser.Email,
				Name:      existingUser.Name,
				CreatedAt: existingUser.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
				IsAdmin:   req.IsAdmin,
			},
		})
		return
	}

	// Create new user
	createParams := &sqlc.CreateUserParams{
		Email: req.Email,
		Name:  req.Name,
	}
	newUser, err := h.App.Queries.CreateUser(c, createParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Add user to company
	addParams := &sqlc.AddUserToCompanyParams{
		UserID:    newUser.ID,
		CompanyID: companyID.(int32),
		IsAdmin:   sql.NullBool{Bool: req.IsAdmin, Valid: true},
	}
	err = h.App.Queries.AddUserToCompany(c, addParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add user to company"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
		"user": UserResponse{
			ID:        newUser.ID,
			Email:     newUser.Email,
			Name:      newUser.Name,
			CreatedAt: newUser.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			IsAdmin:   req.IsAdmin,
		},
	})
}

// DeleteUser soft deletes a user from the company (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Get user ID from context
	requesterID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get company ID from context
	companyID, ok := c.Get("company_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company context not found"})
		return
	}

	// Get user ID from URL params
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	// Parse user ID
	var targetUserID int32
	if _, err := fmt.Sscanf(userID, "%d", &targetUserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Prevent admin from deleting themselves
	if targetUserID == requesterID.(int32) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}

	// Check if user exists and is in the company
	targetUser, err := h.App.Queries.GetUserByID(c, targetUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	checkParams := &sqlc.CheckUserInCompanyParams{
		UserID:    targetUserID,
		CompanyID: companyID.(int32),
	}
	inCompany, err := h.App.Queries.CheckUserInCompany(c, checkParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check user company membership"})
		return
	}
	if !inCompany {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found in this company"})
		return
	}

	// Soft delete the user
	err = h.App.Queries.SoftDeleteUser(c, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user deleted successfully",
		"user": UserResponse{
			ID:        targetUser.ID,
			Email:     targetUser.Email,
			Name:      targetUser.Name,
			CreatedAt: targetUser.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			IsAdmin:   false,
		},
	})
}
