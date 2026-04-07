package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"gobkd/internal/appctx"
)

func RequestLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		fields := logrus.Fields{
			"request_id": GetRequestID(c),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"latency_ms": time.Since(start).Milliseconds(),
			"client_ip":  c.ClientIP(),
		}

		if userID := appctx.GetString(c, appctx.UserIDKey); userID != "" {
			fields["user_id"] = userID
		}

		entry := log.WithFields(fields)
		switch {
		case c.Writer.Status() >= 500:
			entry.Error("request completed")
		case c.Writer.Status() >= 400:
			entry.Warn("request completed")
		default:
			entry.Info("request completed")
		}
	}
}
