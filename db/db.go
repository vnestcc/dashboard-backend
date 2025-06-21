package db

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/config"
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
	fmt.Println(dsn)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		logrus.Fatalf("failed to connect to database: %v", err)
	}
	logrus.Println("Database connection established")
}
