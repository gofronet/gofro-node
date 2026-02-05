package config

import (
	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type Config struct {
	XrayConfigFile string `env:"XRAY_DEFAULT_CONFIG" envDefault:"xconf/config.json"`
	NodeName       string `env:"NODE_NAME"`
	XrayCorePath   string `env:"XRAY_CORE_PATH" envDefault:"xray/xray"`
	IsDevMode      bool   `env:"DEV_MODE" envDefault:"false"`
}

func Load() (*Config, error) {
	godotenv.Load()

	var config Config
	err := env.Parse(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
