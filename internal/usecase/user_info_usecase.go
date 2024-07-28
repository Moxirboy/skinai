package usecase

import (
	"errors"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/domain"
)

func (u usecase) FillInfo(user dto.UserInfo) (int, error) {
	userInfo := u.f.ParseModelToUserInfo(user)
	id, err := u.repo.CreateInfo(*userInfo)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return 0, err
	}
	return id, nil
}
func (u usecase) UpdateInfo(user dto.UserInfo) (id int, err error) {
	userInfo := u.f.ParseModelToUserInfo(user)
	if !Validator(userInfo.Firstname) {

		id, err = u.repo.UpdateName(*userInfo)
		if err != nil {
			u.bot.SendErrorNotification(err)
			return 0, domain.Err("Coudn`t update Name")
		}
	}
	
	if !Validator(userInfo.Gender) {
		id, err = u.repo.UpdateGender(*userInfo)
		if err != nil {
			u.bot.SendErrorNotification(err)
			return 0, domain.Err("Coudn`t update Gender")
		}
	}
	return id, nil
}
func (u usecase) GetUserInfo(userId int) (user dto.UserInfo, err error) {
	exist, err := u.repo.ExistUserInfo(userId)
	if err != nil || errors.Is(err, domain.ErrCouldNotScan) {
		u.bot.SendErrorNotification(err)
		return user, errors.New("no info")
	}
	if !exist {
		return user, errors.New("no info")
	}
	userInfo, err := u.repo.GetUserInfo(userId)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return user, domain.ErrCouldNotScan
	}
	user = *u.f.ParseUserInfoToModel(userInfo)
	return user, nil
}
func (u usecase) GetName(userId int,Error error) (name string, err error) {
	if Error!=nil{
		u.bot.SendErrorNotification(err)
		return "", errors.New("no info")
	}
	exist, err := u.repo.ExistUserInfo(userId)
	if err != nil || errors.Is(err, domain.ErrCouldNotScan) {
		u.bot.SendErrorNotification(err)
		return "", errors.New("no info")
	}
	if !exist {
		return "", errors.New("no info")
	}
	userInfo, err := u.repo.GetUserInfo(userId)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return "", domain.ErrCouldNotScan
	}
	user := *u.f.ParseUserInfoToModel(userInfo)
	return user.Firstname, nil
}