package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"gobkd/internal/response"
)

type SystemHandler struct {
	db *sql.DB
}

func NewSystemHandler(db *sql.DB) *SystemHandler {
	return &SystemHandler{db: db}
}

func (h *SystemHandler) Ping(c *gin.Context) {
	response.OK(c, http.StatusOK, gin.H{"message": "pong"})
}

func (h *SystemHandler) Healthz(c *gin.Context) {
	if err := h.db.PingContext(c.Request.Context()); err != nil {
		response.ServiceUnavailable(c, "database unavailable")
		return
	}

	response.OK(c, http.StatusOK, gin.H{
		"status": "ok",
		"db":     "ok",
	})
}
