package adminconfig

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AdminConfig struct {
	Port           int
	Username       string
	PasswordHash   string
	JWTSecret      string
	TokenExpire    int
	SecureCookie   bool
	TrustedProxies []string
	DBDsn          string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
}

func LoadFromEnv(logger *zap.Logger) *AdminConfig {
	port := 8781
	if v := os.Getenv("ADMIN_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	password := os.Getenv("ADMIN_PASSWORD")
	var passwordHash string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Fatal("failed to hash admin password", zap.Error(err))
		}
		passwordHash = string(hash)
	}

	jwtSecret := os.Getenv("ADMIN_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = generateRandomSecret(32, logger)
		logger.Warn("ADMIN_JWT_SECRET not configured, using random secret — sessions will not survive restart")
	}

	tokenExpire := 24
	if v := os.Getenv("ADMIN_TOKEN_EXPIRE"); v != "" {
		if h, err := strconv.Atoi(v); err == nil {
			tokenExpire = h
		}
	}

	secureCookie := true
	if v := os.Getenv("ADMIN_SECURE_COOKIE"); v == "false" || v == "0" {
		secureCookie = false
	}

	var trustedProxies []string
	if v := os.Getenv("ADMIN_TRUSTED_PROXIES"); v != "" {
		for _, p := range strings.Split(v, ",") {
			if p = strings.TrimSpace(p); p != "" {
				trustedProxies = append(trustedProxies, p)
			}
		}
	}

	return &AdminConfig{
		Port:           port,
		Username:       os.Getenv("ADMIN_USERNAME"),
		PasswordHash:   passwordHash,
		JWTSecret:      jwtSecret,
		TokenExpire:    tokenExpire,
		SecureCookie:   secureCookie,
		TrustedProxies: trustedProxies,
		DBDsn:          os.Getenv("SPEECH_DB_DSN"),
		ReadTimeout:    parseDurationEnv("ADMIN_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:   parseDurationEnv("ADMIN_WRITE_TIMEOUT", 60*time.Second),
		IdleTimeout:    parseDurationEnv("ADMIN_IDLE_TIMEOUT", 120*time.Second),
	}
}

func generateRandomSecret(n int, logger *zap.Logger) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		logger.Fatal("CSPRNG failure")
	}
	return hex.EncodeToString(b)
}

func parseDurationEnv(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return defaultVal
	}
	return d
}
