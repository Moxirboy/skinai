package postgres
const (

	GetAllTitle=`
	SELECT id ,title, owner, createdAt,image FROM news ORDER BY id DESC LIMIT 10 OFFSET $1
	`
	GetById=`
	SELECT  * FROM 	news where id=$1
	`
	GetPicture=`
	select path from news_photo where owner_id=$1 
	`
	NewsCount=`
	select count(id) from news where 1=1 
	`
)