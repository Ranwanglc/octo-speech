package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Mininglamp-OSS/octo-speech/internal/store"
)

func AuthMiddleware(appStore *store.AppStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": http.StatusUnauthorized,
				"msg":    "missing authorization header",
			})
			c.Abort()
			return
		}

		apiKey := strings.TrimPrefix(auth, "Bearer ")
		if apiKey == auth {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": http.StatusUnauthorized,
				"msg":    "invalid authorization format, expected: Bearer <api_key>",
			})
			c.Abort()
			return
		}

		info, err := appStore.Authenticate(apiKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    "authentication error",
			})
			c.Abort()
			return
		}

		if info == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": http.StatusUnauthorized,
				"msg":    "invalid api key",
			})
			c.Abort()
			return
		}

		if info.Status == 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"status": http.StatusForbidden,
				"msg":    "application is disabled",
			})
			c.Abort()
			return
		}

		c.Set("app_id", info.AppID)
		c.Next()
	}
}
