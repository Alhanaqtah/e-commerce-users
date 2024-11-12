package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ENV        string     `yaml:"ENV" env-default:"dev"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Database   Database   `yaml:"postgres"`
	Cache      Cache      `yaml:"redis"`
}

type HTTPServer struct {
	Host        string        `yaml:"host"`
	Port        string        `yaml:"port"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type Database struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"db_name"`
	MaxConns string `yaml:"max_conns"`
}

type Cache struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func MustLoad() *Config {
	var cfg Config

	err := cleanenv.ReadConfig("config/config.yaml", &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
