package domain


type Schedule struct{
	Id string `json:"id"`
	UserId string 
	Name string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Date string `json:"date"`
	FromTime string `json:"from_time"`
	ToTime string `json:"to_time"`
	Purpose string `json:"purpose"`
	Agree string `json:"agree"`
	IsAgree bool `json:"is_agree"`
	DoctorID string `json:"doctor_id"`
	Overview string 
}
type ScheduleResponse struct{
	Id string `json:"id"`
	Name string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Date string `json:"date"`
	FromTime string `json:"from_time"`
	ToTime string `json:"to_time"`
	Purpose string `json:"purpose"`
	IsAgree bool `json:"is_agree"`
	Doctor Doctor
}
