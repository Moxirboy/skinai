package configs

import (
	"github.com/caarlos0/env/v6"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	AppName    string `env:"APP_NAME"`
	AppVersion string `env:"APP_VERSION"`
	Postgres
	JWT
	InitConfig
	BotConfig
	Cookie
	Ai
}
type Postgres struct {
	Port     string `env:"POSTGRES_PORT"`
	Host     string `env:"POSTGRES_HOST"`
	Password string `env:"POSTGRES_PASSWORD"`
	User     string `env:"POSTGRES_USER"`
	Database string `env:"POSTGRES_DATABASE"`
	SslMode  string `env:"POSTGRES_SSLMODE"`
}
type JWT struct {
	SigningKey string `env:"SIGNING"`
	Salt       string `env:"SALT"`
	TokenTTL   string `env:"TOKEN_TTL"`
}
type Server struct {
	Port string
}
type InitConfig struct {
	RunPort string `env:"RUN_PORT"`
}
type BotConfig struct {
	BotToken string `env:"BOT_TOKEN"`
}
type Cookie struct {
	Sessions string `env:"SESSIONS"`
	Token    string `env:"COOKIE_TOKEN"`
	Age      int    `env:"COOKIE_AGE"`
}

type Ai struct {
	Instruction string `env:"INSTRUCTION"`
}

var instance Config

func Configuration() *Config {

	if err := env.Parse(&instance); err != nil {
		panic(err)
	}
	return &instance

}
