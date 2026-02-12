package usecase

import (
	"errors"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/domain"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (u usecase) RegisterUser(newUser *dto.User) (int, error) {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, errors.New("could not hash password")
	}
	newUser.Password = string(hashedPassword)

	user := u.f.CreateUser(newUser)
	id, err := u.repo.Register(*user)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return 0, err
	}
	err = u.repo.CreatePoint(id)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return 0, err
	}
	return id, nil
}
func (u usecase) Exist(newUser dto.User) (bool, error) {
	exist, err := u.repo.Exist(newUser.Username)
	if errors.Is(err, domain.ErrPhoneNumberExist) || !exist {
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
	if !exist {
		return false, 0, nil
	}

	id, hashedPassword, err := u.repo.GetByUsername(user.Username)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return false, 0, err
	}

	// Compare password with bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if err != nil {
		// Fallback: try plain text comparison for old accounts
		if user.Password == hashedPassword {
			// Upgrade to bcrypt hash
			go u.upgradePassword(id, user.Password)
			return true, id, nil
		}
		return false, 0, nil
	}
	return true, id, nil
}

// upgradePassword silently upgrades a plain-text password to bcrypt
func (u usecase) upgradePassword(userID int, plainPassword string) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	u.repo.UpdatePassword(userID, string(hashed))
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

func (u usecase) IsPremium(userId interface{}) (int, error) {
	return u.repo.IsPremium(userId)
}
func (u usecase) UpdatePremium(userId interface{}) error {
	return u.repo.UpdatePremium(userId)
}
func (u usecase) GetPoint(userID interface{}) (value int, err error) {
	return u.repo.GetPoint(userID)
}
