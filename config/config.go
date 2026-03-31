package config

import (
	"fmt"
	"log"
	"net"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	StorageTypeMemory   = "memory"
	StorageTypePostgres = "postgres"
)

type Config struct {
	LoggerLevel       string `env:"LOGGER_LEVEL,required"`
	ServerHost        string `env:"SERVER_HOST,required"`
	ServerPort        string `env:"SERVER_PORT,required"`
	StorageType       string `env:"STORAGE_TYPE,required"`
	PostgresDSN       string `env:"PG_DSN"`
	PlaygroundEnabled bool   `env:"PLAYGROUND_ENABLED" envDefault:"true"`
}

func MustLoad() *Config {
	godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	if err := cfg.validate(); err != nil {
		log.Fatalf("failed to validate config: %v", err)
	}

	return cfg
}

func (cfg *Config) validate() error {
	switch cfg.StorageType {
	case StorageTypeMemory:
		return nil
	case StorageTypePostgres:
		if cfg.PostgresDSN == "" {
			return fmt.Errorf("PG_DSN is required when STORAGE_TYPE=postgres")
		}
		return nil
	default:
		return fmt.Errorf("unsupported STORAGE_TYPE: %s", cfg.StorageType)
	}
}

func (cfg *Config) ServerAddr() string {
	return net.JoinHostPort(cfg.ServerHost, cfg.ServerPort)
}
