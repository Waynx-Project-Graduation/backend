package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/kemit/trip-planner/internal/utils"
)

// ErrorHandler is a global recovery middleware for panics
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v\n%s", err, debug.Stack())
				c.JSON(http.StatusInternalServerError, utils.Response{
					Success: false,
					Error: &utils.ErrorBody{
						Code:    "INTERNAL_ERROR",
						Message: "An unexpected error occurred",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
