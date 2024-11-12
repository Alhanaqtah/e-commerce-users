package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ENV        string     `yaml:"ENV" env-default:"dev"`
	HTTPServer HTTPServer `yaml:"http_server" env-required:"true"`
	Postgres   Postgres   `yaml:"postgres" env-required:"true"`
	Redis      Redis      `yaml:"redis" env-required:"true"`
}

type HTTPServer struct {
	Host        string        `yaml:"host" env-required:"true"`
	Port        string        `yaml:"port" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

type Postgres struct {
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	DBName   string `yaml:"db_name" env-required:"true"`
	MaxConns string `yaml:"max_conns" env-required:"true"`
}

type Redis struct {
	Address  string `yaml:"address" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DB       int    `yaml:"db" env-default:"0"`
}

func MustLoad() *Config {
	var cfg Config

	err := cleanenv.ReadConfig("config/config.yaml", &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
