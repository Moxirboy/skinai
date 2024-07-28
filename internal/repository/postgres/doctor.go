package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testDeployment/internal/domain"
	"testDeployment/internal/repository"
	"testDeployment/pkg/Bot"
)

type doctor struct{
	db *sql.DB
	bot Bot.Bot
}
func NewDoctorRepository(db *sql.DB,bot Bot.Bot) repository.IDoctorRepository{
	return &doctor{
		db:db,
		bot:bot,
	}
}
func (r *doctor) GetAllDoctor(ctx context.Context)([]*domain.DoctorByType,error){
	 doctors:= []*domain.DoctorByType{}
	rows,err:=r.db.QueryContext(ctx,GetAllTypes)
	if err!=nil{
		r.bot.SendErrorNotification(err)
		return nil,err
	}
	for rows.Next(){
		tip :=&domain.DoctorByType{}
		rows.Scan(
			&tip.Type,
		)
		doctors=append(doctors, tip)
	}
	for i,tip:=range doctors{
		rows,err:=r.db.QueryContext(ctx,GetDoctorByType,tip.Type)
		if err!=nil{
			r.bot.SendErrorNotification(err)
			return nil,err
		}
		
		for rows.Next(){
			 doctor:= &domain.DoctorWithType{}
			rows.Scan(
				&doctor.Id,
				&doctor.Name,
				&doctor.Workplace,
				&doctor.WorkHours,
				&doctor.Photo,
				&doctor.PhoneNumber,
				&doctor.Type,
			)
			
			err=r.db.QueryRowContext(ctx,GetRating,doctor.Id).Scan(&doctor.Rating)
			if err!=nil{
				if errors.Is(err,sql.ErrNoRows){
					continue
				}else{
					r.bot.SendErrorNotification(err)
					return nil,err}
				
			}
		
			doctors[i].Doctor=append(doctors[i].Doctor,doctor)
		}
	}
return doctors ,nil
}
func (r *doctor) GetById(ctx context.Context,name string) ([]*domain.DoctorWithType,error){
	doctors:= []*domain.DoctorWithType{}
	rows,err:=r.db.QueryContext(ctx,GetDoctorById,"%"+name+"%")
	if err!=nil{
		r.bot.SendErrorNotification(err)
		return nil,err
	}
	
	for rows.Next(){
	doctor:=&domain.DoctorWithType{}
	rows.Scan(
		&doctor.Id,
		&doctor.Name,
		&doctor.Workplace,
		&doctor.WorkHours,
		&doctor.Photo,
		&doctor.PhoneNumber,
		&doctor.Type,
	)

	doctors=append(doctors, doctor)
}
	
	return doctors,nil
}