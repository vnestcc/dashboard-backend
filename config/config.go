package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type ServerConfig struct {
	Host        string `toml:"host"`
	Port        int    `toml:"port"`
	Prod        bool   `toml:"production"`
	CORS        string `toml:"cors-url"`
	JWTSecret   string `toml:"jwt-secret"`
	TOTPIssuer  string `toml:"totp-issuer"`
	TokenExpiry int    `toml:"token-expiry"`
}

type DBConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	DBName   string `toml:"dbname"`
	SSL      bool   `toml:"ssl"`
}

type Config struct {
	Server ServerConfig `toml:"server"`
	DB     DBConfig     `toml:"db"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		fmt.Println("The config file is invalid")
		return Config{}, err
	}
	return cfg, nil
}
