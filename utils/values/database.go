package values

import "gorm.io/gorm"

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func SetDB(d *gorm.DB) {
	db = d
}
