package handlers

import (
	"github.com/gin-gonic/gin"
)

type pongResponse struct {
	Msg string `json:"msg" example:"pong"`
}

// PingHandler godoc
// @Summary      Health Check
// @Description  Responds with "pong" to indicate the server is alive.
// @Tags         healthcheck
// @Produce      json
// @Success      200  {object}  pongResponse
// @Router       /ping [get]
func PingHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"msg": "pong"})
}
