package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testDeployment/internal/domain"
)
// GetAllDrug is a member function of the struct `repo`. It retrieves all the drug records from the database. The function returns a slice of 
// `domain.Drug` objects and an error. If an error is encountered during the operation (like SQL query execution or data scanning), it sends an 
// error notification, stops the function flow, and returns the error for proper error handling by the calling function. If there are no errors, 
// the function returns all the retrieved drug records and nil as error.

func ( r repo) GetAllDrug()(drugs []domain.Drug,err error){
	query:=`
	select id ,name,description,manufacturer,reciept from drug 
	`
	rows,err:=r.db.Query(query)
	if err!=nil{
		r.Bot.SendErrorNotification(err)
		return nil, err
	}
	for rows.Next() {
		var drug domain.Drug
		err = rows.Scan(
			&drug.Id,
			&drug.Name,
			&drug.Description,
			&drug.Manufacturer,
			&drug.Receipt,
		)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return nil, domain.ErrCouldNotRetrieveFromDataBase
		}
		drugs = append(drugs, drug)
	}
	query=`
	select type_name from type where drug_id=$1 
	`
	for id,drug:=range drugs{
		rows,err:=r.db.Query(query,drug.Id)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return drugs, nil
		}
		for rows.Next(){
			var tip string
		err = rows.Scan(
			&tip,
		)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return drugs, nil
		}
		drugs[id].Type=append(drugs[id].Type, tip)
		}
	}
	return drugs,nil
}
func (r repo) InsertDrug(drug domain.Drug) (id int, err error) {
	query := `
	insert into drug(name,description,manufacturer,reciept) values($1,$2,$3,$4,$5) returning id
	`
	err = r.db.QueryRow(query, drug.Name, drug.Description, drug.Manufacturer, drug.Receipt,drug.Type).Scan(&id)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return 0, err
	}
	query=`
		insert into type(type_name,drug_id) values($1,$2)
	`
	for _ , tip:=range drug.Type{
		_,err:=r.db.Exec(query,tip,id)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return 0, err
		}
	}
	return id, nil
}

func (r repo) GetDrugByName(name string) (drugs []domain.Drug, err error) {
	query := `
    SELECT id ,name,description,manufacturer,reciept FROM drug WHERE LOWER(name) LIKE LOWER($1)
`

	rows, err := r.db.Query(query, "%"+strings.ToLower(name)+"%")
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return nil, err
	}
	for rows.Next() {
		var drug domain.Drug
		err = rows.Scan(
			&drug.Id,
			&drug.Name,
			&drug.Description,
			&drug.Manufacturer,
			&drug.Receipt,
		)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return nil, domain.ErrCouldNotRetrieveFromDataBase
		}
		drugs = append(drugs, drug)
	}
	query=`
	select type_name from type where drug_id=$1 
	`
	for id,drug:=range drugs{
		rows,err:=r.db.Query(query,drug.Id)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return drugs, nil
		}
		for rows.Next(){
			var tip string
		err = rows.Scan(
			&tip,
		)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return drugs, nil
		}
		drugs[id].Type=append(drugs[id].Type, tip)
		}
	}
	return drugs, nil
}
func (r repo) GetDrugById(id string) (drug domain.Drug, err error) {
	query := `
    SELECT id ,name,description,manufacturer,reciept FROM drug WHERE id=$1
`

	err = r.db.QueryRow(query, id).Scan(
		&drug.Id,
		&drug.Name,
		&drug.Description,
		&drug.Manufacturer,
		&drug.Receipt,
	)
	if err != nil {
		r.Bot.SendErrorNotification(err)
		return drug, err
	}
	query=`
	select type_name from type where drug_id=$1 
	`
		rows,err:=r.db.Query(query,drug.Id)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return drug, nil
		}
		for rows.Next(){
			var tip string
		err = rows.Scan(
			&tip,
		)
		if err != nil {
			r.Bot.SendErrorNotification(err)
			return drug, nil
		}
		drug.Type=append(drug.Type, tip)
		}

	return drug, nil
}
func (r *repo) GetDrugByType(ctx context.Context,tip string)(drugs []domain.DrugWithoutType,err error){
	query:=`
select drug_id from type where LOWER(type_name)like LOWER($1)
`
	rows,err:=r.db.QueryContext(ctx,query,"%"+strings.ToLower(tip)+"%")
	if err!=nil{
		if errors.Is(err,sql.ErrNoRows){
			return nil, sql.ErrNoRows
		}
		r.Bot.SendErrorNotification(err)
		return nil, err
	}
	
	var ids []string
	for rows.Next(){
		var id string
		rows.Scan(&id)
		ids=append(ids,id)
	}
	DrugSelect:=`
select id ,name,description,manufacturer,reciept from drug where id=$1
`
	photoSelect:=`
	select path from photos where owner_id=$1
`
	var drug domain.DrugWithoutType
	for _,id:=range ids{
		err = r.db.QueryRow(DrugSelect, id).Scan(
			&drug.Id,
			&drug.Name,
			&drug.Description,
			&drug.Manufacturer,
			&drug.Receipt,
			)
		if err!=nil{
			if errors.Is(err , sql.ErrNoRows){
				return nil, nil
			}
			return nil, err
		}
		rows,err:=r.db.QueryContext(ctx,photoSelect,id)
		if err!=nil{
			if errors.Is(err , sql.ErrNoRows){
				return nil, nil
			}
			return nil, err
		}
		for rows.Next(){
			var photo string
			rows.Scan(&photo)
			drug.Photo=append(drug.Photo,photo)
		}
		drugs=append(drugs,drug)
	}
	return drugs,nil
}
func (r *repo) GetAllTypes(ctx context.Context)(Types []domain.DrugByType,err error){
	query:=`
	SELECT DISTINCT type_name FROM type
	`
	rows,err:=r.db.QueryContext(ctx,query)
	if err!=nil{
		if errors.Is(err,sql.ErrNoRows){
			return nil, nil
		}
		return nil, err
	}
	for rows.Next(){
		var Type domain.DrugByType
		rows.Scan(&Type.Type)
		Types=append(Types,Type)
	}
	SelectID :=`
	select drug_id from type where lower(type_name)=lower($1)
`
	DrugSelect:=`
select id ,name,description,manufacturer,reciept from drug where id=$1
`
	photoSelect:=`
select path from photos where owner_id=$1
`
	for i,Type:=range Types{
		var ids []string
		rows,err:=r.db.QueryContext(ctx,SelectID,Type.Type)
		if err!=nil{
			if errors.Is(err,sql.ErrNoRows){
				return nil, nil
			}
			r.Bot.SendErrorNotification(err)
			return nil, err
		}
		for rows.Next(){
			var id string
			rows.Scan(&id)
			ids=append(ids,id)
		}
		var drug domain.Drug
		for _,id:=range ids{
			err = r.db.QueryRow(DrugSelect, id).Scan(
				&drug.Id,
				&drug.Name,
				&drug.Description,
				&drug.Manufacturer,
				&drug.Receipt,
				)
			if err!=nil{
				if errors.Is(err , sql.ErrNoRows){
					return nil, nil
				}
				return nil, err
			}
			typeSelect:=`
			select type_name from type where drug_id=$1
			`
			row,err:=r.db.QueryContext(ctx,typeSelect,id)
			if err!=nil{
				if errors.Is(err , sql.ErrNoRows){
					return nil, nil
				}
				return nil, err
			}
			for row.Next(){
				var typo string
				row.Scan(&typo)
				drug.Type=append(drug.Type, typo)
			}
			rows,err:=r.db.QueryContext(ctx,photoSelect,id)
			if err!=nil{
				if errors.Is(err , sql.ErrNoRows){
					return nil, nil
				}
				return nil, err
			}
			for rows.Next(){
				var photo string
				rows.Scan(&photo)
				drug.Photo=append(drug.Photo,photo)
			}
			Types[i].Drugs=append(Types[i].Drugs,drug)
		}
	}
	return Types, nil
}