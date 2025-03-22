package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/assylzhan-a/subscription-service/configs"
	"github.com/assylzhan-a/subscription-service/internal/app/auth"
	"github.com/assylzhan-a/subscription-service/internal/app/product"
	"github.com/assylzhan-a/subscription-service/internal/app/subscription"
	"github.com/assylzhan-a/subscription-service/internal/app/voucher"
	"github.com/assylzhan-a/subscription-service/internal/middleware"
	"github.com/assylzhan-a/subscription-service/internal/repository/migrations"
	"github.com/assylzhan-a/subscription-service/internal/repository/postgres"
	httpTransport "github.com/assylzhan-a/subscription-service/internal/transport/http"
	"github.com/assylzhan-a/subscription-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(config.Server.Mode)

	// Connect to database
	db, err := sql.Open("postgres", config.Database.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ping database to verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Run migrations
	if err := migrations.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	productRepo := postgres.NewProductRepository(db)
	subscriptionRepo := postgres.NewSubscriptionRepository(db)
	voucherRepo := postgres.NewVoucherRepository(db)

	// Initialize JWT manager
	jwtManager := jwt.NewManager(config.JWT.SecretKey, config.JWT.Issuer)

	// Initialize auth middleware
	middleware.InitAuthMiddleware(jwtManager)

	// Initialize services
	authService := auth.NewService(userRepo, jwtManager, config.JWT.GetJWTExpirationDuration())
	productService := product.NewService(productRepo)
	subscriptionService := subscription.NewService(subscriptionRepo, productRepo, voucherRepo)
	voucherService := voucher.NewService(voucherRepo, productRepo)

	// Initialize HTTP router
	router := httpTransport.NewRouter(authService, productService, subscriptionService, voucherService)
	router.Setup()

	// Start HTTP server
	address := fmt.Sprintf(":%s", config.Server.Port)
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, router.Engine()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
