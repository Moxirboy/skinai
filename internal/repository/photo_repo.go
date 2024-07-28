package repository


func (r repo) CreatePhoto(id int, path []string) (err error) {
	query := `
	insert into photos(path ,owner_id) values($1,$2) 
	`
	for _, p := range path {
		_, err = r.db.Exec(query, p, id)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return err
		}
	}
	return nil
}
func ( r repo) GetPhotoPath(id int) (path []string,err error){
	query:=`
		select path from photos where owner_id=$1
	`
	rows,err:=r.db.Query(query,id)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return nil,err
	}
	for rows.Next(){
		var p string
		rows.Scan(&p)
		path = append(path, p)
	}
	return path ,nil
}