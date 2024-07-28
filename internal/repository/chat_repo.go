package repository

import (
	"testDeployment/internal/domain"

)



func (r repo ) CreateMessage(userId string,isAi bool,message string,time string) (id int,err error){
query:=`
		insert into messages(user_id ,is_ai,message,created_at) values($1,$2,$3,$4) returning id
`
	err=r.db.QueryRow(query,userId,isAi,message,time).Scan(&id)
	if err!=nil{
		return 0,err
	}
return id,nil
}
func ( r repo) GetAllMessages(userId string )(messages []domain.Message,err error){
	query:=`
		select * from messages where user_id=$1 
	`
	rows,err:=r.db.Query(query,userId)
	if err!=nil{
			r.Bot.SendErrorNotification(err)
			return nil,err
	}
	for rows.Next(){
		var message domain.Message
		rows.Scan(
			&message.Id,
			&message.User_id,
			&message.IsAi,
			&message.Text,
			&message.CreatedAt,
		)
		messages=append(messages, message)
	}
	return messages,nil
}