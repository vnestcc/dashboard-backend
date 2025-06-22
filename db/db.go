package db

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/config"
	"github.com/vnestcc/dashboard/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.DBName,
		cfg.DB.Port,
		func() string {
			if cfg.DB.SSL {
				return "require"
			}
			return "disable"
		}(),
	)
	logger_level := func() logger.LogLevel {
		if cfg.Server.Prod {
			return logger.Silent
		} else {
			return logger.Info
		}
	}
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger_level()),
	})
	if err != nil {
		logrus.Fatalf("failed to connect to database: %v", err)
	}
	logrus.Println("Database connection established")
	DB.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Quarter{},
		&models.FinancialHealth{},
		&models.MarketTraction{},
		&models.UnitEconomics{},
		&models.TeamPerformance{},
		&models.FundraisingStatus{},
		&models.CompetitiveLandscape{},
		&models.OperationalEfficiency{},
		&models.RiskManagement{},
		&models.AdditionalInfo{},
		&models.SelfAssessment{},
		&models.Attachment{},
	)
}
