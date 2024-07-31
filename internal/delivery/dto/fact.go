package dto

type Fact struct {
	Id               int    `json:"id"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	NumberOfQuestion int    `json:"number_of_question"`
}

type FactQuestions struct {
	ID       int       `json:"id"`
	FactId   int       `json:"fact_id"`
	Question string    `json:"question"`
	Choices  []Choices `json:"choices"`
}

type Choices struct {
	Content string `json:"content"`
	IsTrue  bool   `json:"is_true"`
}
