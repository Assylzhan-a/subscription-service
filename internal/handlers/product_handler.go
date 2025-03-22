package handlers

import (
	"net/http"

	"github.com/assylzhan-a/subscription-service/internal/app/product"
	"github.com/assylzhan-a/subscription-service/internal/domain/errors"
	"github.com/assylzhan-a/subscription-service/internal/transport/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productService *product.Service
}

func NewProductHandler(productService *product.Service) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/products", h.GetAllProducts)
	router.GET("/products/:id", h.GetProductByID)
	router.POST("/products", h.CreateProduct)
	router.PUT("/products/:id", h.UpdateProduct)
	router.DELETE("/products/:id", h.DeleteProduct)
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	products, err := h.productService.GetAllProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MapProductsToResponse(products))
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	product, err := h.productService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		if err == errors.ErrProductNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MapProductToResponse(product))
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := product.CreateProductInput{
		Name:           req.Name,
		Description:    req.Description,
		Price:          req.Price,
		DurationMonths: req.DurationMonths,
		TaxRate:        req.TaxRate,
		IsActive:       req.IsActive,
	}

	createdProduct, err := h.productService.CreateProduct(c.Request.Context(), input)
	if err != nil {
		if validationErrors, ok := err.(errors.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrors})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.MapProductToResponse(createdProduct))
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := product.UpdateProductInput{
		ID:             id,
		Name:           req.Name,
		Description:    req.Description,
		Price:          req.Price,
		DurationMonths: req.DurationMonths,
		TaxRate:        req.TaxRate,
		IsActive:       req.IsActive,
	}

	updatedProduct, err := h.productService.UpdateProduct(c.Request.Context(), input)
	if err != nil {
		if err == errors.ErrProductNotFound {
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

	c.JSON(http.StatusOK, dto.MapProductToResponse(updatedProduct))
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	if err := h.productService.DeleteProduct(c.Request.Context(), id); err != nil {
		if err == errors.ErrProductNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
