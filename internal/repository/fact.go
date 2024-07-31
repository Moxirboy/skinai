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
}
