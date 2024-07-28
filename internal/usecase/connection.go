package usecase

import (
	"database/sql"
	repo "testDeployment/internal/repository"
	"testDeployment/internal/repository/postgres"
	"testDeployment/pkg/Bot"
)

type IUseCase interface{
	IOtherUseCase() Usecase
	INewsUsecase() INewsUseCase
	IDoctorUseCase() IDoctorUsecase
	IScheduleUseCase() IScheduleUseCase
}
type SUsecase struct{
	connection map[string]interface{}
}
const (
	_UseCase="Use_Case"
	_NewsUseCase="news_use_case"
	_DoctorUseCase="doctor_use_case"
	_ScheduleUseCase="schedule_use_case"
)
func New(
	db *sql.DB,
	bot Bot.Bot,
) IUseCase  {
	var connections=make(map[string]interface{})
	connections[_UseCase]=NewUserUsecase(
		repo.NewRepo(db,
			bot),
		bot,
	)
	connections[_NewsUseCase]=NewNewsUseCase(
		postgres.NewNewsRepo(db,
		bot,
		),
		bot,
	)
	connections[_DoctorUseCase]=NewDoctorUseCase(
		postgres.NewDoctorRepository(
			db,
			bot,
		),
		bot,
	)
	connections[_ScheduleUseCase]=NewScheduleRepo(
		postgres.NewSchedule(
			db,
			bot,
		),
		bot,
	)
	return &SUsecase{
		connection: connections,
	}
}
func (c *SUsecase) IOtherUseCase() Usecase {
	return c.connection[_UseCase].(Usecase)
}
func (c *SUsecase) INewsUsecase() INewsUseCase {
	return c.connection[_NewsUseCase].(INewsUseCase)
}
func (c *SUsecase) IDoctorUseCase() IDoctorUsecase {
	return c.connection[_DoctorUseCase].(IDoctorUsecase)

}
func (c *SUsecase) IScheduleUseCase() IScheduleUseCase{
	return c.connection[_ScheduleUseCase].(IScheduleUseCase)
}