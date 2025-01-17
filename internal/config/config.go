package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Env              string `env:"ENV" env-default:"local"`
	ConnectionString string `env:"CONNECTION_STRING,required"`
	HTTPServer
}

type HTTPServer struct {
	Address     string        `env:"HTTP_SERVER_ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		panic("error loading .env file")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic("failed to parse environment variables: " + err.Error())
	}

	return &cfg
}
