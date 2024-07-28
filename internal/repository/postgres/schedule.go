package postgres

import (
	"context"
	"database/sql"
	"testDeployment/internal/domain"
	"testDeployment/internal/repository"
	"testDeployment/pkg/Bot"
)

type schedule struct{
	db *sql.DB
	bot Bot.Bot
}
func NewSchedule(db *sql.DB,bot Bot.Bot) repository.IScheduleRepository{
	return &schedule{
		db: db,
		bot: bot,
	}
}
func (r *schedule) Create(ctx context.Context,schedule *domain.Schedule) error{
	_,err:=r.db.ExecContext(
		ctx,
		CreateSchedule,
		schedule.Name,
		schedule.PhoneNumber,
		schedule.Date,
		schedule.FromTime,
		schedule.ToTime,
		schedule.Purpose,
		schedule.IsAgree,
		schedule.Overview,
		schedule.DoctorID,
		schedule.UserId,
	)
	if err!=nil{
		r.bot.SendErrorNotification(err)
		return err
	}
	return nil
}
func(r *schedule ) GetAll(ctx context.Context,userId int)(schedules []*domain.ScheduleResponse,err error){
	rows,err:=r.db.QueryContext(
		ctx,
		GetSchedule,
		userId,
	)
	if err!=nil{
		r.bot.SendErrorNotification(err)
		return nil,err
	}
	for rows.Next(){
		schedule:=&domain.ScheduleResponse{}
		rows.Scan(
			&schedule.Id,
			&schedule.Name,
			&schedule.PhoneNumber,
			&schedule.Date,
			&schedule.FromTime,
			&schedule.ToTime,
			&schedule.Purpose,
			&schedule.IsAgree,
			&schedule.Doctor.Id,
			&schedule.Doctor.Name,
			&schedule.Doctor.Workplace,
			&schedule.Doctor.WorkHours,
			&schedule.Doctor.Photo,
			&schedule.Doctor.PhoneNumber,
		)
		schedules=append(schedules, schedule)
	}
	return schedules,nil
}