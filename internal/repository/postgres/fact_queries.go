package postgres

const (
	createFact     = `INSERT INTO facts (title, content,number_question) VALUES ($1, $2,$3) returning id`
	createQuestion = `insert into question(fact_id,question) values($1,$2) returning id`
	createChoices  = `insert into choices(question_id,content,is_true) values($1,$2,$3) `

	GetFact     = `select id, title,content from facts`
	GetQuestion = `select id,question from question where fact_id=$1 limit 1 OFFSET $2 `
	GetChoices  = `select id,content,is_true from choices where question_id=$1`
)
