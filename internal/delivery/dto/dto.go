package dto

import "time"

type User struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type UserInfo struct {
	Id        int
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	SkinColor int       `json:"skin_color"`
	SkinType  int       `json:"skin_type"`
	Gender    string    `json:"gender"`
	Date      time.Time `json:"date"`
}
