package domain

import (
	"testDeployment/internal/delivery/dto"
	"time"
)

type Factory struct {
}

func (f Factory) CreateUser(newUser *dto.User) *User {
	return &User{
		phone_number: newUser.Email,
		username:     newUser.Username,
		password:     newUser.Password,
		role:         "user",
		created_at:   time.Now().UTC(),
		updated_at:   time.Now().UTC(),
		deleted_at:   nil,
	}
}
func (f Factory) CreateDoctor(newUser *dto.User) *User {
	return &User{
		phone_number: newUser.Email,
		role:         "doctor",
		created_at:   time.Now().UTC(),
		updated_at:   time.Now().UTC(),
		deleted_at:   nil,
	}
}
func (f Factory) ParseModelToDomain(id int, phoneNumber string, role string, createdAt time.Time, updatedAt time.Time, deletedAt *time.Time) dto.User {
	return dto.User{
		Email: phoneNumber,
	}
}
func (f Factory) ParseDomainToModel(u User) {

}
func (f Factory) ParseModelToUserInfo(u dto.UserInfo) *UserInfo {
	return &UserInfo{
		Id:        u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Gender:    u.Gender,
		SkinColor: u.SkinColor,
		SkinType:  u.SkinType,
		UpdatedAt: time.Now(),
		Date:      u.Date,
	}
}
func (f Factory) ParseUserInfoToModel(u UserInfo) *dto.UserInfo {
	return &dto.UserInfo{
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Gender:    u.Gender,
		SkinColor: u.SkinColor,
		SkinType:  u.SkinType,
		Date:      u.Date,
	}
}
