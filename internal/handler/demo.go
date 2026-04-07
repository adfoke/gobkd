package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gobkd/internal/requestx"
	"gobkd/internal/response"
)

type DemoHandler struct{}

func NewDemoHandler() *DemoHandler {
	return &DemoHandler{}
}

type EchoRequest struct {
	Message string `json:"message" binding:"required,max=200"`
}

func (h *DemoHandler) Echo(c *gin.Context) {
	var req EchoRequest
	if !requestx.BindJSON(c, &req) {
		return
	}

	response.OK(c, http.StatusOK, gin.H{
		"message": req.Message,
	})
}
