package usecase

import (
	"context"
	"testDeployment/internal/repository"
	"testDeployment/pkg/Bot"
	"testDeployment/internal/domain"
	"testDeployment/pkg/utils"
)
type newsUseCase struct{
	repo repository.INews
	bot Bot.Bot
}
func NewNewsUseCase(repo repository.INews,bot Bot.Bot) INewsUseCase{
	return &newsUseCase{
		repo: repo,
		bot: bot,
	}
}
func (u newsUseCase)GetAll(ctx context.Context,query utils.PaginationQuery)(news *domain.NewsList,err error){
	news,err=u.repo.GetAll(ctx,query)
	if err!=nil{
		u.bot.SendErrorNotification(err)
		return nil,nil
	}
	return  news,nil
}
func (u newsUseCase) GetOneById(ctx context.Context,id string) (new *domain.NewWithSinglePhoto,err error){
	new,err=u.repo.GetOneById(ctx,id)
	if err!=nil{
		u.bot.SendErrorNotification(err)
		return nil,nil
	}
	return new,nil
}