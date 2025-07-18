package handlers

import (
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/vnestcc/dashboard/config"
	"github.com/vnestcc/dashboard/models"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

type Claims = middleware.Claims

var ResetPasswordCache *cacher.Cacher[string, models.User]

func InitHandler(cfg *config.Config) {
	ResetPasswordCache = cacher.NewCacher[string, models.User](&cacher.NewCacherOpts{
		TimeToLive:    time.Duration(cfg.Server.TokenExpiry) * time.Minute,
		CleanInterval: 2 * time.Hour,
		CleanerMode:   cacher.CleaningCentral,
	})
}
