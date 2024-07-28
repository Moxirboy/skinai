package domain

type NewsList struct {
	TotalCount int `json:"totat_count"`
	TotalPages int `json:"total_pages"`
	Page int `json:"page"`
	Size int `json:"size"`
	HasMore bool `json:"has_more"`
	News []*NewWithSinglePhoto
}
type NewWithSinglePhoto struct{
	ID string `json:"id"`
	Title string `json:"title"`
	Body string `json:"body"`
	Owner string `json:"owner"`
	Photo string `json:"photo"`
	CreatedAt string `json:"created_at"`
}