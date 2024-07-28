package domain

import "time"

type UserInfo struct {
	Id        int
	Firstname string
	Lastname  string
	SkinColor int
	SkinType  int
	Gender    string
	UpdatedAt time.Time
}
