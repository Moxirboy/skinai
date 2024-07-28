package dto

type Response struct {
	Data []Item `json:"data"`
}

type Item struct {
	ID            int    `json:"id"`
	CategoryID    int    `json:"category_id"`
	Date          string `json:"date"`
	Title         string `json:"title"`
	Anons         string `json:"anons"`
	Views         int    `json:"views"`
	AnonsImage    string `json:"anons_image"`
	CategoryTitle string `json:"category_title"`
	CategoryCode  string `json:"category_code"`
	ActivityTitle string `json:"activity_title"`
	ActivityCode  string `json:"activity_code"`
	UrlToWeb      string `json:"url_to_web"`
}
