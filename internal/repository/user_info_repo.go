package repository

import (
	"errors"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/domain"
	"time"
)

func (r repo) UpdateUserInfoDeleted(id int, deleteAt time.Time) (err error) {
	query := `
	update user_info set deleted_at=$1 where user_id=$2
	`
	_, err = r.db.Exec(query, deleteAt, id)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return errors.New("could not delete")
	}
	return nil
}

func (r repo) ExistUserInfo(userId int) (exist bool, err error) {
	query := `
	Select Exists (
		SELECT true
		FROM user_info
		WHERE user_id = $1)
	`
	err = r.db.QueryRow(query, userId).Scan(&exist)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return false, err
	}
	return exist, nil
}
func (r repo) CreateInfo(user domain.UserInfo) (id int, err error) {
	query := `
	insert into  user_info (user_id,firstname,lastname,skin_color,skin_type,gender,created_at,birth) values($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id
`
	row := r.db.QueryRow(query, user.Id, user.Firstname, user.Lastname, user.SkinColor, user.SkinType, user.Gender, user.UpdatedAt, user.Date)
	if err = row.Scan(&id); err != nil {
		r.Bot.SendErrorNotification(err)
		return 0, err
	}
	return id, nil
}
func (r repo) GetUserInfo(userId int) (user domain.UserInfo, err error) {
	query := `
	select id, firstname,lastname,skin_color,skin_type,gender,birth from user_info where user_id=$1
	`
	err = r.db.QueryRow(query, userId).Scan(
		&user.Id,
		&user.Firstname,
		&user.Lastname,
		&user.SkinColor,
		&user.SkinType,
		&user.Gender,
		&user.Date,
	)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return user, domain.ErrCouldNotScan
	}
	return user, nil
}

func (r repo) UpdateInfo(user domain.UserInfo) (id int, err error) {
	query := `
	update user_info set firstname=$2,lastname=$3,skin_color=$4,skin_type=$5,gender=$6,updated_at=$7 where user_id=$1 returning id
	`
	err = r.db.QueryRow(query, user.Id, user.Firstname, user.Lastname, user.SkinColor, user.SkinType, user.Gender, user.UpdatedAt).Scan(&id)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return 0, domain.ErrCouldNotScan
	}
	return id, nil
}
func (r repo) UpdateName(user domain.UserInfo) (id int, err error) {
	query := `
	update user_info set firstname=$2,lastname=$3,updated_at=$4 where user_id=$1 returning id
	`
	err = r.db.QueryRow(query, user.Id, user.Firstname, user.Lastname, user.UpdatedAt).Scan(&id)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return 0, domain.ErrCouldNotScan
	}
	return id, nil
}

func (r repo) UpdateGender(user domain.UserInfo) (id int, err error) {
	query := `
	update user_info set gender=$2,updated_at=$3 where user_id=$1 returning id
	`
	err = r.db.QueryRow(query, user.Id, user.Gender, user.UpdatedAt).Scan(&id)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return 0, domain.ErrCouldNotScan
	}
	return id, nil
}

func (r repo) UpdateEmail(user dto.UserEmail) (id int, err error) {
	query := `
	update users set email=$2 where id=$1 returning id
	`
	err = r.db.QueryRow(query, user.ID, user.Email).Scan(&id)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return 0, domain.ErrCouldNotScan
	}
	return id, nil
}
