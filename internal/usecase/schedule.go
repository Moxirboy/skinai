package usecase

import (
	"testDeployment/internal/repository"
	"testDeployment/pkg/Bot"
	"context"
	"testDeployment/internal/domain"
)

type schedule struct{
	repo repository.IScheduleRepository
	bot Bot.Bot
}
func NewScheduleRepo(repo repository.IScheduleRepository,bot Bot.Bot) IScheduleUseCase{
	return &schedule{
		repo: repo,
		bot: bot,
	}
}
func (u *schedule )Create(ctx context.Context,schedule *domain.Schedule) error{
	return u.repo.Create(ctx,schedule)
}
func(u *schedule )GetAll(ctx context.Context,userId int)([]*domain.ScheduleResponse,error){
	return u.repo.GetAll(ctx,userId)
}