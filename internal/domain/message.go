package domain


type Message struct{
	Id int `json:"id"`
	User_id string`json:"user_id"`
	IsAi bool`json:"is_AI"`
	Text string`json:"message"`
	CreatedAt string`json:"sent_at"`
}
type NewMessage struct{
	Request string `json:"request"`
	UserId string `json:"id"`
	Name string `json:"name"`
}
type Response struct{
	Response string `json:"request"`
}
