package company

import (
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/vnestcc/dashboard/models"
	middleware "github.com/vnestcc/dashboard/utils/middlewares"
)

type Claims = middleware.Claims

var StartupCache = cacher.NewCacher[uint, models.Company](&cacher.NewCacherOpts{
	TimeToLive:    2 * time.Minute,
	CleanInterval: 1 * time.Hour,
	Revaluate:     true,
	CleanerMode:   cacher.CleaningCentral,
})

var QuarterCache = cacher.NewCacher[string, models.Quarter](&cacher.NewCacherOpts{
	TimeToLive:    3 * time.Minute,
	CleanInterval: 1 * time.Hour,
	Revaluate:     true,
})

type createCompanyRequest struct {
	Name         string `json:"name" binding:"required" example:"Acme Inc"`
	ContactName  string `json:"contact_name" binding:"required" example:"John Doe"`
	ContactEmail string `json:"contact_email" binding:"required,email" example:"john@acme.com"`
	Sector       string `json:"sector" binding:"required" example:"xyz"`
	Description  string `json:"description" binding:"required" example:"We do something xyz and make money"`
}

type quarterResponse struct {
	ID      uint   `json:"id" example:"1"`
	Quarter string `json:"quarter" example:"Q1"`
	Year    uint   `json:"year" example:"2025"`
	Date    string `json:"date,omitempty" example:"2025-04-01T00:00:00Z"`
}

type joinCompanyRequest struct {
	SecretCode string `json:"secret_code" binding:"required" example:"random hex"`
}

type quarterRequest struct {
	Quarter string `json:"quarter" binding:"required" example:"Q1"`
	Year    uint   `json:"year" binding:"required" example:"2024"`
}

type versionInfo struct {
	Version    uint32
	IsEditable uint16
}
