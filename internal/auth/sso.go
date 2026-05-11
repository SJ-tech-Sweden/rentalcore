package auth

import (
	"errors"
	"os"
	"strings"
	"time"

	"go-barcode-webapp/internal/models"

	"github.com/golang-jwt/jwt/v4"
)

const minSigningKeyLength = 32

// Claims contains the fields encoded into the SSO JWT.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// getSigningKey returns the HMAC secret used for SSO tokens.
// Falls back to RENTALCORE encryption key when SSO secret is not set.
func getSigningKey() ([]byte, error) {
	if k := strings.TrimSpace(os.Getenv("SSO_JWT_SECRET")); k != "" {
		if len(k) < minSigningKeyLength {
			return nil, errors.New("SSO_JWT_SECRET must be at least 32 characters")
		}
		return []byte(k), nil
	}
	k := strings.TrimSpace(os.Getenv("ENCRYPTION_KEY"))
	if k == "" {
		return nil, errors.New("missing SSO_JWT_SECRET or ENCRYPTION_KEY")
	}
	if len(k) < minSigningKeyLength {
		return nil, errors.New("ENCRYPTION_KEY must be at least 32 characters")
	}
	return []byte(k), nil
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
	signingKey, err := getSigningKey()
	if err != nil {
		return "", err
	}
	signed, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return signed, nil
}

// VerifySSOToken validates the token and returns claims if valid.
func VerifySSOToken(tokenStr string) (*Claims, error) {
	signingKey, err := getSigningKey()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
