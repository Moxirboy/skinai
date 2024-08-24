package dto

type User struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password" format:"password"`
}
type UserInfo struct {
	Id        int
	Firstname string `json:"firstname" example:"Uyg'un'"`
	Lastname  string `json:"lastname"`
	SkinColor int    `json:"skin_color"`
	SkinType  int    `json:"skin_type"`
	Gender    string `json:"gender" example:"transgender"`
	Date      string `json:"date" example:"2006-02-23"`
}

type UserEmail struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}
