package http

import (
	"github.com/assylzhan-a/subscription-service/internal/app/product"
	"github.com/assylzhan-a/subscription-service/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine         *gin.Engine
	productService *product.Service
}

func NewRouter(
	productService *product.Service,
) *Router {
	return &Router{
		engine:         gin.Default(),
		productService: productService,
	}
}

func (r *Router) Setup() {
	// CORS middleware
	r.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	v1 := r.engine.Group("/api/v1")

	productHandler := handlers.NewProductHandler(r.productService)

	productHandler.RegisterRoutes(v1)

	// Health check
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
