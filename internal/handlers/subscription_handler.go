package handlers

import (
	"net/http"

	"github.com/assylzhan-a/subscription-service/internal/app/subscription"
	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/middleware"
	"github.com/assylzhan-a/subscription-service/internal/transport/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	subscriptionService *subscription.Service
}

func NewSubscriptionHandler(subscriptionService *subscription.Service) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

func (h *SubscriptionHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Protected routes
	subscriptionRouter := router.Group("")
	subscriptionRouter.Use(middleware.GetAuthMiddleware().Authenticate())
	{
		subscriptionRouter.POST("", h.CreateSubscription)
		subscriptionRouter.GET("", h.GetUserSubscriptions)
		subscriptionRouter.GET("/:id", h.GetSubscriptionByID)
		subscriptionRouter.PATCH("/:id/pause", h.PauseSubscription)
		subscriptionRouter.PATCH("/:id/unpause", h.UnpauseSubscription)
		subscriptionRouter.PATCH("/:id/cancel", h.CancelSubscription)
	}
}

func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req dto.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	input := subscription.CreateSubscriptionInput{
		UserID:    userID,
		ProductID: productID,
		WithTrial: req.WithTrial,
	}

	createdSubscription, err := h.subscriptionService.CreateSubscription(c.Request.Context(), input)
	if err != nil {
		if validationErrors, ok := err.(errors.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrors})
			return
		}
		if err == errors.ErrProductNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == errors.ErrInactiveProduct {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.MapSubscriptionToResponse(createdSubscription))
}

func (h *SubscriptionHandler) GetUserSubscriptions(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	subscriptions, err := h.subscriptionService.GetUserSubscriptions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MapSubscriptionsToResponse(subscriptions))
}

func (h *SubscriptionHandler) GetSubscriptionByID(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	subscription, err := h.subscriptionService.GetSubscriptionByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrSubscriptionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Ensure the subscription belongs to the authenticated user
	if subscription.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, dto.MapSubscriptionToResponse(subscription))
}

func (h *SubscriptionHandler) PauseSubscription(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	// Verify ownership
	subscription, err := h.subscriptionService.GetSubscriptionByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrSubscriptionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if subscription.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.subscriptionService.PauseSubscription(c.Request.Context(), id); err != nil {
		if err == errors.ErrSubscriptionNotActive {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == errors.ErrSubscriptionInTrial {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription paused successfully"})
}

func (h *SubscriptionHandler) UnpauseSubscription(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	// Verify ownership
	subscription, err := h.subscriptionService.GetSubscriptionByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrSubscriptionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if subscription.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.subscriptionService.UnpauseSubscription(c.Request.Context(), id); err != nil {
		if err == errors.ErrSubscriptionAlreadyPaused {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription unpaused successfully"})
}

func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	// Verify ownership
	subscription, err := h.subscriptionService.GetSubscriptionByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrSubscriptionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if subscription.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.subscriptionService.CancelSubscription(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription cancelled successfully"})
}
