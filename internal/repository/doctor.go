package repository
import (
	"context"
	"testDeployment/internal/domain"
)
type IDoctorRepository interface{
	GetAllDoctor(ctx context.Context)([]*domain.DoctorByType,error)
	GetById(ctx context.Context,name string) ([]*domain.DoctorWithType,error)
}