package auth

import (
	"strings"
	"testing"

	"go-barcode-webapp/internal/models"
)

func TestGenerateSSOToken_ValidatesTTLBounds(t *testing.T) {
	t.Setenv("SSO_JWT_SECRET", strings.Repeat("x", 32))

	u := &models.User{UserID: 1, Username: "tester"}

	if _, err := GenerateSSOToken(u, 0); err == nil {
		t.Fatal("expected error for ttlSeconds=0")
	}
	if _, err := GenerateSSOToken(u, -1); err == nil {
		t.Fatal("expected error for negative ttlSeconds")
	}
	if _, err := GenerateSSOToken(u, maxSSOTokenTTLSeconds+1); err == nil {
		t.Fatal("expected error for ttlSeconds above max")
	}
	if _, err := GenerateSSOToken(u, maxSSOTokenTTLSeconds); err != nil {
		t.Fatalf("expected success at ttl upper bound, got error: %v", err)
	}
}
