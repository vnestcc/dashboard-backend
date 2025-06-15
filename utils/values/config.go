package values

import "github.com/vnestcc/dashboard/config"

var cfg *config.Config

func SetConfig(c *config.Config) {
	cfg = c
}

func GetConfig() *config.Config {
	return cfg
}
