package configs

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	_ "github.com/mattn/go-sqlite3"
)

func NewPostgresConfig() (*sql.DB, error) {
	instance, err := sql.Open("pgx","postgresql://postgres:wwgUZgtjPIEhJlxUsqGevVcMozwqkrMB@roundhouse.proxy.rlwy.net:48249/railway")
	if err != nil {
		panic(err)
	}
	err = instance.Ping()
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
