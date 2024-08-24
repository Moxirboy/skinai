package dto

type User struct {
	Email    string `json:"email" `
	Username string `json:"username"`
	Password string `json:"password" format:"password"`
}
type UserInfo struct {
	Id        int
	Firstname string `json:"firstname" example:"Uyg'un'"`
	Lastname  string `json:"lastname" example:"Tursunov"`
	SkinColor int    `json:"skin_color" example:"0"`
	SkinType  int    `json:"skin_type" example:"0"`
	Gender    string `json:"gender" example:"male"`
	Date      string `json:"date" example:"2005-05-22"`
}

type UserEmail struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}
