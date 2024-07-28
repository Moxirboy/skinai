package Bot

import (
	"fmt"
	"log"
	"runtime"
	"time"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type bot struct {
	*tgbotapi.BotAPI
}
type Bot interface {
	SendErrorNotification(err error)
	SendNotification(mess string)
}

func NewBot(botAPI *tgbotapi.BotAPI) Bot {
	return &bot{BotAPI: botAPI}
}

const chatID = int64(-4103413678)

func (b bot) SendErrorNotification(err error) {
	// Replace "USER_CHAT_ID" with the actual user's chat ID to send the notification
	if err==nil{
		return
	}
	_, file, line, _ := runtime.Caller(1) // Get information about the calling function (1 level up in the call stack)
	message := fmt.Sprintf(time.Now().Format("2006/01/02  15:04:05\n")+"Error in %s:%d\n%v", file, line, err)
	msg := tgbotapi.NewMessage(chatID, message+err.Error())
	_, err = b.Send(msg)
	if err != nil {
		log.Printf("Error sending notification to user: %v", err)
	}

}
func (b bot) SendNotification(message string) {
	// Replace "USER_CHAT_ID" with the actual user's chat ID to send the notification
	_, file, line, _ := runtime.Caller(1) // Get information about the calling function (1 level up in the call stack)
	logEntry := fmt.Sprintf(time.Now().Format("2006/01/02  15:04:05 \nmessage: ")+"[%s:%d]\n%s", file, line, message)
	msg := tgbotapi.NewMessage(chatID, logEntry)
	_, err := b.Send(msg)
	if err != nil {
		log.Printf("Error sending notification to user: %v", err)
	}
}
