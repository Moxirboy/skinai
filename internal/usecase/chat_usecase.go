package usecase

import (
	"sort"
	"strconv"
	"testDeployment/internal/domain"
	"time"
)
func(u usecase) SaveMessage(userId string,isAi bool,message string) (id int,err error){
	time:=time.Now().Format("2006-01-02T15:04:05")
	u.bot.SendNotification(time)
	id ,err=u.repo.CreateMessage(userId,isAi,message,time)
	if err!=nil{
		u.bot.SendErrorNotification(err)
		return id, err
	}
	return id ,nil
}
func (u usecase) GetAllMessages(userId int )(messages []domain.Message,err error){

	messages,err=u.repo.GetAllMessages(strconv.Itoa(userId))
	if err!=nil{
		u.bot.SendErrorNotification(err)
		return nil, err
	}

	sort.Slice(messages, func(i, j int) bool {
		parseTime := func(s string) time.Time {

			layout := "2006-01-02T15:04:05"
			
			t, err := time.Parse(layout, s)
			if err != nil {
				u.bot.SendErrorNotification(err)
			}
			return t
		}
	
		return parseTime(messages[i].CreatedAt).Before(parseTime(messages[j].CreatedAt))
	})
	

	return messages,nil
}