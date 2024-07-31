package repository

import (
	"context"
	"database/sql"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/domain"
	"testDeployment/pkg/Bot"
	"time"
)

type userDB struct {
	id           int
	phone_number string
	role         string
	created_at   time.Time
	updated_at   time.Time
	deleted_at   *time.Time
}
type repo struct {
	db  *sql.DB
	f   domain.Factory
	Bot Bot.Bot
}
type Repo interface {
	Register(user domain.User) (int, error)
	Exist(email string) (bool, error)
	GetByUsername(username string) (id int, password string, err error)
	GetAll() []dto.User
	UpdatePhoneNumber(number string) (int, error)
	UpdateIsActive(id int, deleteAt time.Time) (err error)
	UpdateUserInfoDeleted(id int, deleteAt time.Time) (err error)
	CreateInfo(user domain.UserInfo) (int, error)
	GetUserInfo(userId int) (domain.UserInfo, error)
	ExistUserInfo(userId int) (bool, error)
	UpdateInfo(user domain.UserInfo) (id int, err error)
	UpdateName(user domain.UserInfo) (id int, err error)
	UpdateGender(user domain.UserInfo) (id int, err error)
	UpdateVerified(userId interface{}) (err error)
	InsertDrug(drug domain.Drug) (id int, err error)
	CreatePhoto(id int, path []string) (err error)
	GetDrugByName(name string) (drugs []domain.Drug, err error)
	GetDrugById(id string) (drug domain.Drug, err error)
	GetAllDrug() (drugs []domain.Drug, err error)
	CreateDoctorInfo(info domain.Doctor) (id int, err error)
	GetPhotoPath(id int) (path []string, err error)
	CreateMessage(userId string, isAi bool, message string, time string) (id int, err error)
	GetAllMessages(userId string) (messages []domain.Message, err error)
	UpdatePhoto(path string) (id int, err error)
	GetDrugByType(ctx context.Context, tip string) (drugs []domain.DrugWithoutType, err error)
	GetAllTypes(ctx context.Context) (Types []domain.DrugByType, err error)
	IsPremium(userId interface{}) (int, error)
	UpdatePremium(userId interface{}) error
	CreatePoint(userId interface{}) (err error)
	GetPoint(userID interface{}) (value int, err error)
}

func NewRepo(db *sql.DB, bot Bot.Bot) Repo {
	return &repo{db: db,
		Bot: bot,
	}
}
