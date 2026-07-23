package middleware

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidAccessToken = errors.New("invalid or expired access token")

// ParseAccessToken validates an access-token JWT (signature + expiry, via
// jwt/v5's own default claim validation) and returns its claims. Shared by
// AuthRequired (REST) and the realtime WebSocket/SSE handshake handlers
// (internal/realtime/websocket_handler.go) — consolidated from two
// previously-duplicated jwt.Parse call sites so both agree on the same
// signing secret and validation rules, which matters now that the access
// token TTL is short (15 min) rather than 24h.
func ParseAccessToken(tokenValue string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenValue, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidAccessToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidAccessToken
	}
	return claims, nil
}

// UserIDFromClaims extracts user_id from already-parsed claims.
func UserIDFromClaims(claims jwt.MapClaims) string {
	userID, _ := claims["user_id"].(string)
	return userID
}
