package usecase

import (
	"errors"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/domain"
	"time"
)

func (u usecase) RegisterUser(newUser *dto.User) (int, error) {

	user := u.f.CreateUser(newUser)
	id, err := u.repo.Register(*user)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return 0, err
	}
	return id, nil
}
func (u usecase) Exist(newUser dto.User) (bool, error) {
	exist, err := u.repo.Exist(newUser.Email)
	if errors.Is(err, domain.ErrPhoneNumberExist) || !exist {
		u.bot.SendErrorNotification(err)
		return false, nil
	}
	return exist, nil
}
func (u usecase) UpdateIsVerified(userId interface{}) (err error) {
	err = u.repo.UpdateVerified(userId)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return err
	}
	return nil
}

func (u usecase) Login(user dto.User) (bool, int, error) {
	exist, err := u.repo.Exist(user.Username)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return false, 0, err
	}
	if exist {
		id, password, err := u.repo.GetByUsername(user.Username)
		if err != nil {
			u.bot.SendErrorNotification(err)
			return false, 0, err
		}
		match := user.Password == password
		return match, id, nil
	} else {
		return false, 0, nil
	}
}
func (u usecase) GetAll() (User []dto.User) {
	return u.repo.GetAll()
}
func (u usecase) DeleteUser(id int) (err error) {
	deletedAt := time.Now()
	err = u.repo.UpdateIsActive(id, deletedAt)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return err
	}
	err = u.repo.UpdateUserInfoDeleted(id, deletedAt)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return err
	}
	return nil
}

func (u usecase) IsPremium(userId interface{}) (int,error){
	return u.repo.IsPremium(userId)
}
func (u usecase) UpdatePremium(userId interface{}) (error){
	return u.UpdatePremium(userId)
}