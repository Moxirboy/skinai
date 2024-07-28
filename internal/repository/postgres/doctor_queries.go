package postgres


const (
	GetAllTypes=`
		select distinct type from doctor 
	`
	GetDoctorByType=`
	select id, name,workplace, work_hours,photo,phone_number,type from doctor where type=$1
	`
	GetDoctorById=`
		select id,name,workplace, work_hours,photo ,phone_number,type from doctor where name like $1 
	`
	GetRating=`
	SELECT 
    CASE 
        WHEN COUNT(id) > 0 THEN CAST(SUM(score) AS DECIMAL) / CAST(COUNT(id) AS DECIMAL)
        ELSE 0 
    END AS average_score
FROM scores
WHERE user_id = $1;
 `

)
