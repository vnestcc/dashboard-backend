package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vnestcc/dashboard/utils/values"
)

// pongResponse defines the structure for ping response
type pongResponse struct {
	Msg string `json:"msg" example:"pong"`
}

// healthcheckResponse defines the structure for healthcheck response
type healthcheckResponse struct {
	Status   string `json:"status" example:"ok"`
	Database string `json:"database" example:"ok"`
}

// PingHandler godoc
// @Summary      Health Check
// @Description  Responds with "pong" to indicate the server is alive.
// @Tags         healthcheck
// @Produce      json
// @Success      200  {object}  pongResponse
// @Router       /ping [get]
func PingHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, pongResponse{Msg: "pong"})
}

// HealthcheckHandler godoc
// @Summary      Health Check (DB)
// @Description  Responds with status and database connectivity check.
// @Tags         healthcheck
// @Produce      json
// @Success      200  {object}  healthcheckResponse
// @Failure      500  {object}  healthcheckResponse
// @Router       /healthcheck [get]
func HealthcheckHandler(ctx *gin.Context) {
	db := values.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, healthcheckResponse{
			Status:   "error",
			Database: "connection_error",
		})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		ctx.JSON(http.StatusInternalServerError, healthcheckResponse{
			Status:   "error",
			Database: "unreachable",
		})
		return
	}
	ctx.JSON(http.StatusOK, healthcheckResponse{
		Status:   "ok",
		Database: "ok",
	})
}
