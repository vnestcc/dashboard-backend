package utils

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils/values"
)

func UserCleanUp() {
	db := values.GetDB()
	now := time.Now()
	cutoff := now.Add(-6 * time.Hour)

	db.Where("role = ? AND startup_id IS NULL AND created_at <= ?", "user", cutoff).Delete(&models.User{})

	logrus.Println("Scheduled cleanup ran at:", now.Format(time.RFC3339))
}
