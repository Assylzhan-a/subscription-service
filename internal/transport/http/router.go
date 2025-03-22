package http

import (
	"github.com/assylzhan-a/subscription-service/internal/app/auth"
	"github.com/assylzhan-a/subscription-service/internal/app/product"
	"github.com/assylzhan-a/subscription-service/internal/app/subscription"
	"github.com/assylzhan-a/subscription-service/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine              *gin.Engine
	authService         *auth.Service
	productService      *product.Service
	subscriptionService *subscription.Service
}

func NewRouter(
	authService *auth.Service,
	productService *product.Service,
	subscriptionService *subscription.Service,
) *Router {
	return &Router{
		engine:              gin.Default(),
		authService:         authService,
		productService:      productService,
		subscriptionService: subscriptionService,
	}
}

func (r *Router) Setup() {
	r.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	v1 := r.engine.Group("/api/v1")

	authHandler := handlers.NewAuthHandler(r.authService)
	productHandler := handlers.NewProductHandler(r.productService)
	subscriptionHandler := handlers.NewSubscriptionHandler(r.subscriptionService)

	authHandler.RegisterRoutes(v1.Group("/auth"))
	productHandler.RegisterRoutes(v1)
	subscriptionHandler.RegisterRoutes(v1.Group("/subscriptions"))

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
