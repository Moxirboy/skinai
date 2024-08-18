package domain

type Message struct {
	Id        int    `json:"id"`
	User_id   string `json:"user_id"`
	IsAi      bool   `json:"is_AI"`
	Text      string `json:"message"`
	CreatedAt string `json:"sent_at"`
}
type NewMessage struct {
	Request string `json:"message"`
}
type Response struct {
	Response string `json:"request"`
}
