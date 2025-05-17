package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type Mode int

const (
	Dev Mode = iota
	Prod
)

func (m Mode) String() string {
	if m == Prod {
		return "prod"
	}
	return "dev"
}

func getMode() Mode {
	mode := os.Getenv("MODE")
	if mode == "prod" {
		return Prod
	}
	return Dev
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	URL string
}

type Config struct {
	Mode           Mode
	DatabaseConfig *DatabaseConfig
	RedisConfig    *RedisConfig
}

func New() (*Config, error) {
	if getMode() == Dev {
		err := godotenv.Load()
		if err != nil {
			return nil, err
		}
	}
	return &Config{
		DatabaseConfig: &DatabaseConfig{
			DSN: os.Getenv("DATABASE_DSN"),
		},
		RedisConfig: &RedisConfig{
			URL: os.Getenv("REDIS_URL"),
		},
		Mode: getMode(),
	}, nil
}
