package dto

type User struct {
	Email    string `json:"email"`
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

// ── Auth Response DTOs ──

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	UserID      int    `json:"user_id,omitempty"`
	Role        string `json:"role"`
	ExpiresIn   int    `json:"expires_in"` // seconds
}

type GuestResponse struct {
	AccessToken  string `json:"access_token"`
	Role         string `json:"role"`
	AILimit      int    `json:"ai_limit"`
	UploadLimit  int    `json:"upload_limit"`
	ExpiresIn    int    `json:"expires_in"`
	Message      string `json:"message"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=6"`
}
