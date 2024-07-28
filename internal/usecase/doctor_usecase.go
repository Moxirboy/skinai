package usecase

import (
	"errors"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/domain"
)

func (u usecase) RegisterDoctor(newUser *dto.User) (int, error) {
	exist, err := u.repo.Exist(newUser.Email)
	if errors.Is(err, domain.ErrPhoneNumberExist) || exist {
		u.bot.SendErrorNotification(err)
		return 0, domain.ErrPhoneNumberExist
	}
	doctor := u.f.CreateDoctor(newUser)
	return u.repo.Register(*doctor)
}
func (u usecase) FillDoctorInfo(info domain.Doctor) (int ,error){
	empty:=isAnyFieldEmpty(info)
	if empty{
		return 0 ,domain.ErrEmptyField
	}
	id,err:=u.repo.CreateDoctorInfo(info)
	if err!=nil{
		return 0 ,err
	}
	return id ,err
}
func (u usecase) ChangeInfo(info domain.Doctor)(id int,err error){
	id,err=u.repo.CreateDoctorInfo(info)
	if err!=nil{
		u.bot.SendErrorNotification(err)
		return 0, err
	}
	return id, nil
}
func (u usecase) ChangePhoto(path string) (id int ,err error){
	id,err=u.repo.UpdatePhoto(path)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return 0, err
	}
	return id, nil
}
//func(u usecase) PublishRecommStress()
//// func (u usecase) ChangeInfo()
//// func (u usecase) ChangePhoto(photo string) (id  int ,err error){
//
//// }