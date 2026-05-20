package admin

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignVerifyJWT_RoundTrip(t *testing.T) {
	secret := "round-trip-secret"
	username := "testuser"

	token, expiresAt, err := signJWT(username, secret, 1)
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if time.Until(expiresAt) < 50*time.Minute {
		t.Errorf("expected expiry ~1h from now, got %v", expiresAt)
	}

	claims, err := verifyJWT(token, secret)
	if err != nil {
		t.Fatalf("verifyJWT failed: %v", err)
	}
	if claims.Username != username {
		t.Errorf("expected username %q, got %q", username, claims.Username)
	}
}

func TestVerifyJWT_ExpiredToken(t *testing.T) {
	secret := "expire-secret"
	claims := &Claims{
		Username: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	_, err = verifyJWT(tokenStr, secret)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerifyJWT_WrongSecret(t *testing.T) {
	token, _, err := signJWT("user", "correct-secret", 1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = verifyJWT(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestGenerateCSRFToken_Format(t *testing.T) {
	token := generateCSRFToken()

	if len(token) != 32 {
		t.Errorf("expected 32 hex chars (16 bytes), got length %d", len(token))
	}

	_, err := hex.DecodeString(token)
	if err != nil {
		t.Errorf("expected valid hex string, got error: %v", err)
	}
}

func TestGenerateCSRFToken_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := generateCSRFToken()
		if seen[token] {
			t.Fatalf("duplicate CSRF token: %s", token)
		}
		seen[token] = true
	}
}
