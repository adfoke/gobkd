package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gobkd/internal/apperr"
	authx "gobkd/internal/auth"
	"gobkd/internal/response"
	"gobkd/internal/service"
)

type UserHandler struct {
	auth        *authx.Service
	userService *service.UserService
}

func NewUserHandler(auth *authx.Service, userService *service.UserService) *UserHandler {
	return &UserHandler{
		auth:        auth,
		userService: userService,
	}
}

func (h *UserHandler) Me(c *gin.Context) {
	authUser, err := h.auth.CurrentUser(c.Request)
	if err != nil {
		response.FromError(c, apperr.Unauthorized("login required"))
		return
	}

	user, err := h.userService.SyncCurrentUser(c.Request.Context(), authUser)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, http.StatusOK, gin.H{
		"user": user,
		"auth": authUser,
	})
}
