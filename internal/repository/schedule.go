package repository

import (
	"context"
	"testDeployment/internal/domain"
)

type IScheduleRepository interface{
	Create(ctx context.Context,schedule *domain.Schedule) error
	GetAll(ctx context.Context,userId int)([]*domain.ScheduleResponse,error)
}