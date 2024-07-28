package postgres

const (
	CreateSchedule = `
	insert into schedule(name,phone_number,date,from_time,to_time,purpose,is_agree,overview,doctor_id,user_id) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`
	GetSchedule = `SELECT s.id, s.name, s.phone_number, s.date, s.from_time, s.to_time, s.purpose, s.is_agree,
    d.id AS doctor_id, d.name AS doctor_name, d.workplace, d.work_hours, d.photo, d.phone_number AS doctor_phone_number
FROM schedule s
JOIN doctor d ON CAST( s.doctor_id AS Integer) =d.id 
WHERE CAST(s.user_id AS Integer) =  $1`

)
