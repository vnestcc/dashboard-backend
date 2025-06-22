package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/utils"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start_time := time.Now()
		ctx.Next()
		status := ctx.Writer.Status()
		log := utils.Logger.WithFields(logrus.Fields{
			"status":          status,
			"method":          ctx.Request.Method,
			"path":            ctx.Request.URL.Path,
			"errors":          ctx.Errors.String(),
			"time(nano)":      time.Since(start_time).Nanoseconds(),
			"message_context": ctx.Value("message"),
		})
		switch status {
		case 404:
			log.Trace("Path not found")
		case 500:
			log.Warn("Internal server error")
		default:
			log.Info("Request Processed")
		}
	}
}
