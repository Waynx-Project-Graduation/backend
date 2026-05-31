package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kemit/trip-planner/internal/utils"
)

// AuthMiddleware validates JWT token and sets user context
func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		tokenStr := utils.ExtractToken(authHeader)
		if tokenStr == "" {
			utils.Unauthorized(c, "Invalid authorization format. Use: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenStr)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				utils.Unauthorized(c, "Token has expired")
			} else {
				utils.Unauthorized(c, "Invalid token")
			}
			c.Abort()
			return
		}

		// Set user info in context for downstream handlers
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}
