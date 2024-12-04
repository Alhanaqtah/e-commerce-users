package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	ENV        string     `env:"ENV" env-default:"dev"`
	Prefix     string     `env:"PREFIX" env-default:""`
	HTTPServer HTTPServer `env-required:"true"`
	Postgres   Postgres   `env-required:"true"`
	Redis      Redis      `env-required:"true"`
	SMTP       SMTP       `env-required:"true"`
	Tokens     Tokens     `env-required:"true"`
}

type HTTPServer struct {
	Host        string        `env:"HTTP_SERVER_HOST" env-required:"true"`
	Port        string        `env:"HTTP_SERVER_PORT" env-required:"true"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env-required:"true"`
}

type Postgres struct {
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     string `env:"POSTGRES_PORT" env-required:"true"`
	DBName   string `env:"POSTGRES_NAME" env-required:"true"`
	MaxConns int    `env:"POSTGRES_MAX_CONNS" env-required:"true"`
}

type Redis struct {
	Address  string `env:"REDIS_ADDRESS" env-required:"true"`
	Password string `env:"REDIS_PASSWORD" env-required:"true"`
	DB       int    `env:"REDIS_DB" env-default:"0"`
}

type SMTP struct {
	Username string        `env:"SMTP_USERNAME" env-required:"true"`
	Password string        `env:"SMTP_PASSWORD" env-required:"true"`
	Host     string        `env:"SMTP_HOST" env-required:"true"`
	Port     string        `env:"SMTP_PORT" env-required:"true"`
	CodeTTL  time.Duration `env:"SMTP_CODE_TTL" env-required:"true"`
}

type Tokens struct {
	Secret     string        `env:"TOKENS_SECRET" env-required:"true"`
	AccessTTL  time.Duration `env:"TOKENS_ACCESS_TTL" env-required:"true"`
	RefreshTTL time.Duration `env:"TOKENS_REFRESH_TTL" env-required:"true"`
}

func MustLoad() *Config {
	var cfg Config

	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system environment variables")
	}

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
