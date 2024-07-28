package repository

import (
	"testDeployment/internal/domain"

)
func (r repo) CreateDoctorInfo(info domain.Doctor) (id int, err error){
	query:=`
	insert into doctor (name,workplace,photo) values ($1,$2,$3) returning id	
		`
	err=r.db.QueryRow(query,info.Name,info.Workplace,info.Photo).Scan(&id)
	if err!=nil{
		r.Bot.SendErrorNotification(err)
		return 0,err
	}
	return id,err
}
func(r repo) UpdatePhoto(path string) (id int,err error){
	query:=`
	UPDATE doctor SET photo=$1 RETURNING id
`
	err=r.db.QueryRow(query,path).Scan(&id)
	if err!=nil{
		r.Bot.SendErrorNotification(err)
		return 0, err
	}
	return id, nil
}