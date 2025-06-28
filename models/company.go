package models

import (
	"crypto/rand"
	"encoding/hex"

	"gorm.io/gorm"
)

type Company struct {
	gorm.Model
	ID           uint `gorm:"primaryKey;autoIncrement"`
	Name         string
	ContactName  string
	ContactEmail string `gorm:"unique"`
	SecretCode   string `gorm:"unique"`
	Sector       string
	Description  string

	Quarters []Quarter `gorm:"foreignKey:CompanyID"`
}

func (c *Company) BeforeCreate(tx *gorm.DB) (err error) {
	random := make([]byte, 6) // this much size to avoid collisions
	_, err = rand.Read(random)
	if err != nil {
		return err
	}
	c.SecretCode = hex.EncodeToString(random)
	return nil
}
