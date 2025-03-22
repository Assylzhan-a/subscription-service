package handlers

import (
	"net/http"

	"github.com/assylzhan-a/subscription-service/internal/app/voucher"
	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/domain/models"
	"github.com/assylzhan-a/subscription-service/internal/middleware"
	"github.com/assylzhan-a/subscription-service/internal/transport/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VoucherHandler struct {
	voucherService *voucher.Service
}

func NewVoucherHandler(voucherService *voucher.Service) *VoucherHandler {
	return &VoucherHandler{
		voucherService: voucherService,
	}
}

func (h *VoucherHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes
	publicRouter := router.Group("/vouchers")
	{
		publicRouter.POST("/validate", h.ValidateVoucher)
	}

	// potential admin routes for voucher management
	adminRouter := router.Group("/admin/vouchers")
	adminRouter.Use(middleware.GetAuthMiddleware().Authenticate())
	{
		adminRouter.POST("", h.CreateVoucher)
		adminRouter.GET("", h.GetAllVouchers)
		adminRouter.GET("/:id", h.GetVoucherByID)
		adminRouter.GET("/product/:id", h.GetVouchersByProductID)
		adminRouter.PUT("/:id", h.UpdateVoucher)
		adminRouter.DELETE("/:id", h.DeleteVoucher)
	}
}

func (h *VoucherHandler) ValidateVoucher(c *gin.Context) {
	var req dto.ValidateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	input := voucher.ValidateVoucherInput{
		Code:      req.Code,
		ProductID: productID,
	}

	voucherObj, err := h.voucherService.ValidateVoucher(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusOK, dto.ValidateVoucherResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	response := dto.MapVoucherToResponse(voucherObj)
	c.JSON(http.StatusOK, dto.ValidateVoucherResponse{
		Valid:   true,
		Voucher: &response,
	})
}

func (h *VoucherHandler) CreateVoucher(c *gin.Context) {
	var req dto.CreateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var productID *uuid.UUID
	if req.ProductID != nil {
		id, err := uuid.Parse(*req.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}
		productID = &id
	}

	input := voucher.CreateVoucherInput{
		Code:          req.Code,
		DiscountType:  models.DiscountType(req.DiscountType),
		DiscountValue: req.DiscountValue,
		ProductID:     productID,
		ExpiresAt:     req.ExpiresAt,
		IsActive:      req.IsActive,
	}

	createdVoucher, err := h.voucherService.CreateVoucher(c.Request.Context(), input)
	if err != nil {
		if validationErrors, ok := err.(errors.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrors})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.MapVoucherToResponse(createdVoucher))
}

func (h *VoucherHandler) GetAllVouchers(c *gin.Context) {
	vouchers, err := h.voucherService.GetAllActiveVouchers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MapVouchersToResponse(vouchers))
}

func (h *VoucherHandler) GetVoucherByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	voucherObj, err := h.voucherService.GetVoucherByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrVoucherNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MapVoucherToResponse(voucherObj))
}

func (h *VoucherHandler) GetVouchersByProductID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	vouchers, err := h.voucherService.GetVouchersByProductID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MapVouchersToResponse(vouchers))
}

func (h *VoucherHandler) UpdateVoucher(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	var req dto.UpdateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var productID *uuid.UUID
	if req.ProductID != nil {
		pid, err := uuid.Parse(*req.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}
		productID = &pid
	}

	input := voucher.UpdateVoucherInput{
		ID:            id,
		Code:          req.Code,
		DiscountType:  models.DiscountType(req.DiscountType),
		DiscountValue: req.DiscountValue,
		ProductID:     productID,
		ExpiresAt:     req.ExpiresAt,
		IsActive:      req.IsActive,
	}

	updatedVoucher, err := h.voucherService.UpdateVoucher(c.Request.Context(), input)
	if err != nil {
		if err == errors.ErrVoucherNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if validationErrors, ok := err.(errors.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrors})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MapVoucherToResponse(updatedVoucher))
}

func (h *VoucherHandler) DeleteVoucher(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}

	if err := h.voucherService.DeleteVoucher(c.Request.Context(), id); err != nil {
		if err == errors.ErrVoucherNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
