//go:generate swag init
package main

// @title						VNEST Dashboard Swagger docs
// @version					1.0
// @description			This endpoint is for dev purposes
// @BasePath				/api

// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" and your JWT token for authentication and authorization

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/vnestcc/dashboard/config"
	"github.com/vnestcc/dashboard/db"
	"github.com/vnestcc/dashboard/handlers"
	"github.com/vnestcc/dashboard/routers"
	"github.com/vnestcc/dashboard/utils"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
	"github.com/vnestcc/dashboard/utils/values"
)

func main() {
	var cfg config.Config
	if config, err := config.LoadConfig("/config.toml"); err != nil {
		fmt.Printf("Error in loading config file: %v\n", err)
		return
	} else {
		cfg = config
	}
	values.SetConfig(&cfg)
	db.InitDB(&cfg)
	values.SetDB(db.DB)
	utils.NewLogger(cfg.Server.Prod)
	if cfg.Server.Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	s := gocron.NewScheduler(time.UTC)
	s.Every("6h").Do(utils.UserCleanUp)
	s.StartAsync()
	handlers.InitHandler(&cfg)
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.Server))
	r.Use(gin.Recovery())
	routers.LoadRoutes(r)
	fmt.Printf("[ENGINE] Server started at %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	r.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
}
