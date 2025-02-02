package config

import (
	"time"

	"github.com/Rolan335/Musiclib/internal/repository/postgres"
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type ExternalApiConfig struct {
	URL string `env:"EXTERNAL_API_URL"`
}

type Config struct {
	Port           string        `env:"PORT"`
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT"`
	LogLevel       string        `env:"LOG_LEVEL"`
	GinMode        string        `env:"GIN_MODE"`
	DB             postgres.Config
	API            ExternalApiConfig
	Migration      postgres.MigrationConfig
}

func MustNewConfig() *Config {
	//load env file
	if err := godotenv.Load(".env"); err != nil {
		panic("failed to load env file: " + err.Error())
	}

	//parse env file
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		panic("failed to parse env: " + err.Error())
	}
	return &cfg
}
