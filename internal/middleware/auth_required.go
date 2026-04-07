package middleware

import (
	"github.com/gin-gonic/gin"

	"gobkd/internal/appctx"
	authx "gobkd/internal/auth"
	"gobkd/internal/response"
)

func RequireUser(authService *authx.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := authService.CurrentUser(c.Request)
		if err != nil {
			response.Unauthorized(c, "login required")
			return
		}

		appctx.SetString(c, appctx.UserIDKey, user.ID)
		c.Next()
	}
}
