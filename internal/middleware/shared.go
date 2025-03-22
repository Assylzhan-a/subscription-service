package middleware

import (
	"github.com/assylzhan-a/subscription-service/pkg/jwt"
)

var authMiddleware *AuthMiddleware

// InitAuthMiddleware initializes the global auth middleware
func InitAuthMiddleware(jwtManager *jwt.Manager) {
	authMiddleware = NewAuthMiddleware(jwtManager)
}

func GetAuthMiddleware() *AuthMiddleware {
	return authMiddleware
}
