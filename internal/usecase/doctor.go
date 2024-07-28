package usecase

import (
	"context"
	"testDeployment/internal/domain"
	"testDeployment/internal/repository"
	"testDeployment/pkg/Bot"
)

type doctorUseCase struct{
	repo repository.IDoctorRepository
	bot Bot.Bot
}
func NewDoctorUseCase(repo repository.IDoctorRepository,bot Bot.Bot) IDoctorUsecase{
	return &doctorUseCase{
repo: repo,
bot: bot,
	}
}
func (u *doctorUseCase) GetAll(ctx context.Context) ([]*domain.DoctorByType,error){
	return u.repo.GetAllDoctor(ctx)
}
func (u *doctorUseCase) GetOneByID(ctx context.Context,name string)  ([]*domain.DoctorWithType,error){
	return u.repo.GetById(ctx,name)
}