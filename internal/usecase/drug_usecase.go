package usecase

import (
	"strconv"
	"context"
	"testDeployment/internal/domain"
)

func (u usecase) CreateDrug(drug domain.Drug) (id int, err error) {
	id, err = u.repo.InsertDrug(drug)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return 0, err
	}
	err = u.repo.CreatePhoto(id, drug.Photo)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return 0, err
	}
	return id, nil
}
func (u usecase) GetDrugs(drug domain.DrugSearch) (drugS []domain.Drug, err error) {
	drugS, err = u.repo.GetDrugByName(drug.Name)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return nil, err
	}
	
	for i,drug:=range drugS{
		id,err:=strconv.Atoi(drug.Id)
		if err != nil {
			u.bot.SendErrorNotification(err)
			return nil, err
		}
		drugS[i].Photo,err=u.repo.GetPhotoPath(id)
		if err != nil {
			u.bot.SendErrorNotification(err)
			return nil, err
		}
	}
	return drugS, nil
}

func (u usecase) GetDrug(d domain.DrugSearch) (drug domain.Drug, err error) {
	drug, err = u.repo.GetDrugById(d.Id)
	if err != nil {
		u.bot.SendErrorNotification(err)
		return drug, err
	}
	id,err:=strconv.Atoi(drug.Id)
		if err != nil {
			u.bot.SendErrorNotification(err)
			return drug, err
		}
		drug.Photo,err=u.repo.GetPhotoPath(id)
		if err != nil {
			u.bot.SendErrorNotification(err)
			return drug, err
		}
	return drug, nil
}
func (u usecase) GetAllDrug()(drugs []domain.Drug,err error){
	drugs, err = u.repo.GetAllDrug()
	if err != nil {
		u.bot.SendErrorNotification(err)
		return nil, err
	}
	for i,drug:=range drugs{
		id,err:=strconv.Atoi(drug.Id)
		if err != nil {
			u.bot.SendErrorNotification(err)
			return nil, err
		}
		drugs[i].Photo,err=u.repo.GetPhotoPath(id)
		if err != nil {
			u.bot.SendErrorNotification(err)
			return nil, err
		}
	}
	return drugs, nil
}
func (u usecase)GetDrugByType(ctx context.Context,tip string)(drugs []domain.DrugWithoutType,err error){
	return u.repo.GetDrugByType(ctx,tip)
}
func(u usecase) GetAllTypes(ctx context.Context)(Types []domain.DrugByType,err error){
	return u.repo.GetAllTypes(ctx)
}