package usecase

import (
	"context"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/repository"
	"testDeployment/pkg/Bot"
)

type factUseCase struct {
	repo repository.IFactRepository
	bot  Bot.Bot
}

func NewFactUseCase(repo repository.IFactRepository, bot Bot.Bot) IFactUseCase {
	return &factUseCase{
		repo: repo,
		bot:  bot,
	}
}

func (u factUseCase) CreateQuestion(
	ctx context.Context,
	id int,
	questions *[]dto.FactQuestions,
) error {
	for _, question := range *questions {
		err := u.repo.CreateQuestion(
			ctx,
			id,
			&question)
		if err != nil {
			return err
		}
		err = u.repo.CreateChoices(
			ctx,
			question.ID,
			question.Choices)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u factUseCase) CreateFact(
	ctx context.Context,
	fact *dto.Fact,
) error {
	err := u.repo.CreateFact(
		ctx,
		fact)
	if err != nil {
		return err
	}
	return nil
}

func (u factUseCase) GetFacts(
	ctx context.Context,
) ([]dto.Fact, error) {
	return u.repo.GetFacts(
		ctx)
}
func (u factUseCase) GetQuestion(ctx context.Context, id int, offset int) (*dto.FactQuestions, error) {
	return u.repo.GetQuestion(
		ctx,
		id,
		offset,
	)
}
