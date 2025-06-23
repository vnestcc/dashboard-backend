package models

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/vnestcc/dashboard/utils/values"
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
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memoryCost, timeCost, parallelism, b64Salt, b64Hash)
	u.Password = encoded
	if code, err := generateResetCode(12); err != nil {
		return err
	} else {
		u.BackupCode = code
	}
	u.CreatedAt = time.Now()
	if key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      values.GetConfig().Server.TOTPIssuer,
		AccountName: u.Email,
		Period:      60,
		Digits:      otp.Digits(otp.DigitsEight),
	}); err != nil {
		return err
	} else {
		u.TOTPSecret = key.Secret()
	}
	return
}

func (u *User) SetPassword(password string) (err error) {
	err = nil
	salt := make([]byte, saltLength)
	if _, err = rand.Read(salt); err != nil {
		return
	}
	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, uint8(parallelism), keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memoryCost, timeCost, parallelism, b64Salt, b64Hash)
	u.Password = encoded
	return
}

func (u *User) TOTPUrl() (string, error) {
	issuer := values.GetConfig().Server.TOTPIssuer
	if key, err := otp.NewKeyFromURL(
		fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
			issuer, u.Email, u.TOTPSecret, issuer),
	); err != nil {
		return "", err
	} else {
		return key.URL(), nil
	}
}

func (u *User) VerifyTOTP(otp string) bool {
	return totp.Validate(otp, u.TOTPSecret)
}

func (u *User) ComparePassword(password string) error {
	parts := strings.Split(u.Password, "$")
	if len(parts) != 6 {
		return fmt.Errorf("invalid hash format")
	}
	var m, t, p uint32
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return fmt.Errorf("invalid params section: %v", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("invalid salt encoding: %v", err)
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("invalid hash encoding: %v", err)
	}
	actualHash := argon2.IDKey([]byte(password), salt, t, m, uint8(p), uint32(len(expectedHash)))
	if !bytes.Equal(actualHash, expectedHash) {
		return fmt.Errorf("wrong password")
	}
	return nil
}
