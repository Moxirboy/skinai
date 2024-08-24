package repository

import (
	"context"
	"testDeployment/internal/delivery/dto"
)

type IFactRepository interface {
	CreateChoices(
		ctx context.Context,
		id int,
		choices []dto.Choices,
	) error
	CreateQuestion(
		ctx context.Context,
		id int,
		question *dto.FactQuestions,
	) error
	CreateFact(
		ctx context.Context,
		fact *dto.Fact,
	) error
	GetFacts(
		ctx context.Context,
	) ([]dto.Fact, error)
	GetQuestion(ctx context.Context, id int, offset int) (dto.FactQuestions, error)
	GetChoices(ctx context.Context, id int) ([]dto.Choices, error)
	UpdatePoint(ctx context.Context, id int) (int, error)
	UpdateImage(ctx context.Context, id int, path string) error
}
