package auth

import (
	"errors"
	"os"
	"time"

	"go-barcode-webapp/internal/models"

	"github.com/golang-jwt/jwt/v4"
)

// Claims contains the fields encoded into the SSO JWT.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// getSigningKey returns the HMAC secret used for SSO tokens.
// Falls back to RENTALCORE encryption key when SSO secret is not set.
func getSigningKey() []byte {
	if k := os.Getenv("SSO_JWT_SECRET"); k != "" {
		return []byte(k)
	}
	return []byte(os.Getenv("ENCRYPTION_KEY"))
}

// GenerateSSOToken creates a signed JWT for the given user and TTL seconds.
func GenerateSSOToken(u *models.User, ttlSeconds int) (string, error) {
	if u == nil {
		return "", errors.New("nil user")
	}
	now := time.Now()
	claims := &Claims{
		UserID:   u.UserID,
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttlSeconds) * time.Second)),
			Issuer:    "rentalcore",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(getSigningKey())
	if err != nil {
		return "", err
	}
	return signed, nil
}

// VerifySSOToken validates the token and returns claims if valid.
func VerifySSOToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getSigningKey(), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
