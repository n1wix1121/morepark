package handler

import (
	"net/http"

	"morepark/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки пользователей"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
	IsActive bool   `json:"is_active"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Create(c.Request.Context(), service.CreateUserRequest{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
		Role:     req.Role,
		IsActive: req.IsActive,
	})
	if err != nil {
		if err == service.ErrInvalidRole {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrUserEmailTaken {
			c.JSON(http.StatusConflict, gin.H{"error": "Email уже занят"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

type UpdateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required,min=2"`
	Password string `json:"password"`
	Role     string `json:"role" binding:"required"`
	IsActive bool   `json:"is_active"`
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, service.UpdateUserRequest{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
		Role:     req.Role,
		IsActive: req.IsActive,
	})
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrInvalidRole {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUserID, _ := c.Get("user_id")

	err := h.userService.Delete(c.Request.Context(), id, currentUserID.(string))
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case service.ErrCannotDeleteSelf, service.ErrLastDirector:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользователь удалён"})
}
