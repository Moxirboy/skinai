package configs

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func NewPostgresConfig(config *Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		config.Postgres.User,
		config.Postgres.Password,
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.Database,
	)
	instance, err := sql.Open("pgx", connStr)
	if err != nil {
		panic(err)
	}

	return instance, nil
}

func BotConfi(as string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(as)
	if err != nil {
		log.Println(err)
	}
	return bot, nil
}

func NewBotConfig(cfg Config) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Println(err)
	}
	return bot, nil
}
