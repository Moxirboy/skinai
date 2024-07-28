package postgres

import (
	"context"
	"database/sql"
	"testDeployment/internal/domain"
	"testDeployment/internal/repository"
	"testDeployment/pkg/Bot"
	"testDeployment/pkg/utils"
)


type news struct{
	db *sql.DB
	bot Bot.Bot
}
func NewNewsRepo(db *sql.DB,bot Bot.Bot) repository.INews{
	return &news{
		db: db,
		bot: bot,
	}
}
func (r *news) GetAll(ctx context.Context,query utils.PaginationQuery)( *domain.NewsList, error){
	count := 0
	if err := r.db.QueryRowContext(ctx,
		NewsCount).Scan(
		&count); err != nil {
		r.bot.SendErrorNotification(err)
		return nil, err
	}
	if count == 0 {
		return &domain.NewsList{
			TotalCount: count,
			TotalPages: utils.GetTotalPages(count, query.GetSize()),
			Page:       query.GetPage(),
			Size:       query.GetSize(),
			HasMore:    utils.GetHasMore(query.GetPage(), count, query.GetSize()),
			News:       make([]*domain.NewWithSinglePhoto, 0),
		}, nil
	}

	rows, err := r.db.QueryContext(
		ctx,
		GetAllTitle,
		query.GetOffset(),
	)
	if err != nil {
		r.bot.SendErrorNotification(err)
		return nil, err
	}
	defer rows.Close()
	newsList := make([]*domain.NewWithSinglePhoto, 0, query.GetSize())
	for rows.Next() {
		new := &domain.NewWithSinglePhoto{} 
		rows.Scan(
			&new.ID,
			&new.Title,
			&new.Owner,
			&new.CreatedAt,
			&new.Photo,
		)
		newsList = append(newsList, new)
	}

	return &domain.NewsList{
		TotalCount: count,
		TotalPages: utils.GetTotalPages(count, query.GetSize()),
		Page:       query.GetPage(),
		Size:       query.GetSize(),
		HasMore:    utils.GetHasMore(query.GetPage(), count, query.GetSize()),
		News:       newsList,
	}, nil
}
func (r *news) 	GetOneById(ctx context.Context,id string) (*domain.NewWithSinglePhoto, error){
	news := domain.NewWithSinglePhoto{}
	err := r.db.QueryRowContext(
		ctx,
		GetById,
		id,
	).Scan(
		&news.ID,
		&news.Title,
		&news.Body,
		&news.Owner,
		&news.CreatedAt,
		&news.Photo,
	)
	if err != nil {
		r.bot.SendErrorNotification(err)
		return nil, err
	}
	return &news, nil
}