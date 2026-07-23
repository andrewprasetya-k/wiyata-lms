package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		//cek apakah ada header authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		claims, err := ParseAccessToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user", claims)

		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	userClaims, exists := c.Get("user")
	if !exists {
		return ""
	}

	claims, ok := userClaims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return ""
	}

	return userID
}

func GetEmail(c *gin.Context) string {
	userClaims, exists := c.Get("user")
	if !exists {
		return ""
	}

	claims, ok := userClaims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	email, ok := claims["email"].(string)
	if !ok {
		return ""
	}

	return email
}
