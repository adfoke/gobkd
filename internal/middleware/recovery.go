package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"gobkd/internal/appctx"
	"gobkd/internal/response"
)

func Recovery(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.WithFields(logrus.Fields{
					"request_id": appctx.GetString(c, appctx.RequestIDKey),
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"panic":      fmt.Sprint(rec),
					"stack":      string(debug.Stack()),
				}).Error("panic recovered")

				response.InternalError(c, "internal server error")
			}
		}()

		c.Next()
	}
}
