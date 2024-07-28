package repository

import ( 
	     "context"
         "testDeployment/pkg/utils"
		 "testDeployment/internal/domain"
)
type INews interface{
	GetAll(ctx context.Context,query utils.PaginationQuery)(news *domain.NewsList,err error)
	GetOneById(ctx context.Context,id string) (new *domain.NewWithSinglePhoto,err error)
}