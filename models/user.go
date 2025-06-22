package models

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

const (
	timeCost    = 1
	memoryCost  = 32 * 1024
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

type User struct {
	gorm.Model
	ID         uint `gorm:"primaryKey;autoIncrement"`
	Name       string
	Position   string
	Email      string `gorm:"unique"`
	Password   string
	Role       string `gorm:"not null"`
	Approved   bool   `gorm:"default:false"`
	BackupCode string `gorm:"unique"`
	TOTPSecret string `gorm:"unique"`
	StartupID  *uint
	StartUp    *Company `gorm:"foreignKey:StartupID;references:ID"`
}

func generateResetCode(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += n.String()
	}
	return code, nil
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	err = nil
	salt := make([]byte, saltLength)
	if _, err = rand.Read(salt); err != nil {
		return
	}
	hash := argon2.IDKey([]byte(u.Password), salt, timeCost, memoryCost, uint8(parallelism), keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=19$t=%d$m=%d$p=%d%s%s", timeCost, memoryCost, parallelism, b64Salt, b64Hash)
	u.Password = encoded
	if code, err_ := generateResetCode(12); err_ != nil {
		err = err_
		return
	} else {
		u.BackupCode = code
	}
	u.CreatedAt = time.Now()
	return
}

func (u *User) ComparePassword(password string) (bool, error) {
	parts := strings.Split(u.Password, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}
	var t, m, p uint32
	_, err := fmt.Sscanf(parts[3], "t=%d", &t)
	if err != nil {
		return false, err
	}
	_, err = fmt.Sscanf(parts[3], "m=%d", &m)
	if err != nil {
		return false, err
	}
	_, err = fmt.Sscanf(parts[3], "p=%d", &p)
	if err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	actualHash := argon2.IDKey([]byte(password), salt, t, m, uint8(p), uint32(len(expectedHash)))
	if bytes.Equal(actualHash, expectedHash) {
		return true, nil
	}
	return false, nil
}
