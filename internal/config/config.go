package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ENV        string     `yaml:"ENV" env-default:"dev"`
	App        App        `yaml:"app"`
	HTTPServer HTTPServer `yaml:"http_server" env-required:"true"`
	Postgres   Postgres   `yaml:"postgres" env-required:"true"`
	Redis      Redis      `yaml:"redis" env-required:"true"`
	Tokens     Tokens     `yaml:"tokens" env-required:"true"`
}

type App struct {
	Prefix string `yaml:"prefix"`
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

type Tokens struct {
	Secret     string        `yaml:"secret" env-required:"true"`
	AccessTTL  time.Duration `yaml:"access_ttl" env-required:"true"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env-required:"true"`
}

func MustLoad() *Config {
	var cfg Config

	confPath := os.Getenv("AUTH_CONF_PATH")
	if confPath == "" {
		panic("AUTH_CONF_PATH not found")
	}

	err := cleanenv.ReadConfig(confPath, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
