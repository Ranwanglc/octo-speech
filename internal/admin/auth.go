package admin

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Mininglamp-OSS/octo-speech/internal/adminconfig"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func signJWT(username string, secret string, expireHours int) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(expireHours) * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	return tokenStr, expiresAt, err
}

func verifyJWT(tokenStr string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

func generateCSRFToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatal("CSPRNG failure: ", err)
	}
	return hex.EncodeToString(b)
}

func setAuthCookies(c *gin.Context, token string, csrfToken string, cfg *adminconfig.AdminConfig) {
	maxAge := cfg.TokenExpire * 3600
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    token,
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   cfg.SecureCookie,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   cfg.SecureCookie,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearAuthCookies(c *gin.Context, cfg *adminconfig.AdminConfig) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Secure:   cfg.SecureCookie,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Secure:   cfg.SecureCookie,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})
}
