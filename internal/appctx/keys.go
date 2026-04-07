package appctx

import "github.com/gin-gonic/gin"

const (
	RequestIDKey = "request_id"
	UserIDKey    = "user_id"
)

func SetString(c *gin.Context, key, value string) {
	c.Set(key, value)
}

func GetString(c *gin.Context, key string) string {
	if v, ok := c.Get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
