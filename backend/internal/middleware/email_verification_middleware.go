package middleware

import (
	"backend/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireVerifiedUser(userRepo interface {
	GetByID(id string) (*domain.User, error)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		user, err := userRepo.GetByID(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if user.EmailVerifiedAt == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: email not verified"})
			c.Abort()
			return
		}

		c.Next()
	}
}
