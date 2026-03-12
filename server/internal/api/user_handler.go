package api

import (
	"strconv"

	"com.hermes.platform/internal/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authService services.AuthService
}

func NewUserHandler(authService services.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	users, total, err := h.authService.ListUsers(page, pageSize)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	var userList []gin.H
	for _, user := range users {
		var roles []string
		for _, role := range user.Roles {
			roles = append(roles, role.Name)
		}
		userList = append(userList, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"roles":      roles,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	Success(c, gin.H{
		"users":     userList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.authService.GetUserByID(uint(id))
	if err != nil {
		NotFound(c, "User not found")
		return
	}

	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}

	Success(c, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"roles":      roles,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid user ID")
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	err = h.authService.UpdateUser(uint(id), req.Name)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "User updated successfully"})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid user ID")
		return
	}

	err = h.authService.DeleteUser(uint(id))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) AssignRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid user ID")
		return
	}

	var req struct {
		RoleIDs []uint `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	err = h.authService.AssignRolesToUser(uint(id), req.RoleIDs)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "Roles assigned successfully"})
}
