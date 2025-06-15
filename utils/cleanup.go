package utils

import (
	"fmt"
	"time"

	"github.com/vnestcc/dashboard/models"
	"github.com/vnestcc/dashboard/utils/values"
)

var db = values.GetDB()

func UserCleanUp() {
	now := time.Now()
	cutoff := now.Add(-6 * time.Hour)

	db.Where("role = ? AND startup_id IS NULL AND created_at <= ?", "user", cutoff).
		Delete(&models.User{})

	fmt.Println("Scheduled cleanup ran at:", now.Format(time.RFC3339))
}
