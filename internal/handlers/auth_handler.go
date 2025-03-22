package handlers

import (
	"net/http"

	"github.com/assylzhan-a/subscription-service/internal/app/auth"
	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/middleware"
	"github.com/assylzhan-a/subscription-service/internal/transport/dto"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *auth.Service
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/register", h.RegisterUser)
	router.POST("/login", h.LoginUser)
	router.GET("/me", middleware.GetAuthMiddleware().Authenticate(), h.GetMe)
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req dto.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := auth.RegisterUserInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	user, err := h.authService.RegisterUser(c.Request.Context(), input)
	if err != nil {
		if validationErrors, ok := err.(errors.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrors})
			return
		}
		if err == errors.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.MapUserToResponse(user))
}

func (h *AuthHandler) LoginUser(c *gin.Context) {
	var req dto.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := auth.LoginUserInput{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := h.authService.LoginUser(c.Request.Context(), input)
	if err != nil {
		if validationErrors, ok := err.(errors.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrors})
			return
		}
		if err == errors.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		User:      dto.MapUserToResponse(response.User),
		Token:     response.Token,
		ExpiresAt: response.ExpiresAt,
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err == errors.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MapUserToResponse(user))
}
