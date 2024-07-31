package postgres

import (
	"context"
	"database/sql"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/repository"
	"testDeployment/pkg/Bot"
)

type fact struct {
	db  *sql.DB
	bot Bot.Bot
}

func NewFactRepository(db *sql.DB, bot Bot.Bot) repository.IFactRepository {
	return &fact{
		db:  db,
		bot: bot,
	}
}

func (r fact) CreateFact(
	ctx context.Context,
	fact *dto.Fact,
) error {
	err := r.db.QueryRowContext(ctx, createFact, fact.Title, fact.Content, fact.NumberOfQuestion).Scan(&fact.Id)
	if err != nil {
		return err
	}
	return nil
}

func (r fact) CreateQuestion(
	ctx context.Context,
	id int,
	question *dto.FactQuestions,
) error {
	err := r.db.QueryRowContext(ctx, createQuestion, id, question.Question).Scan(&question.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r fact) CreateChoices(
	ctx context.Context,
	id int,
	choices []dto.Choices,
) error {
	for _, choice := range choices {
		_, err := r.db.ExecContext(ctx, createChoices, id, choice.Content, choice.IsTrue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r fact) GetFacts(ctx context.Context) ([]dto.Fact, error) {
	row, err := r.db.QueryContext(
		ctx,
		GetFacts,
	)
	if err != nil {
		row.Close()
		return nil, err
	}
	facts := []dto.Fact{}
	for row.Next() {
		var fact dto.Fact
		err := row.Scan(
			&fact.Id,
			&fact.Title,
			&fact.Content,
			&fact.NumberOfQuestion,
		)
		if err != nil {
			row.Close()
			return nil, err
		}
		facts = append(facts, fact)
	}
	row.Close()
	return facts, nil
}
