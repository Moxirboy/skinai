package delivery

import (
	request "testDeployment/internal/delivery/http"
	"testDeployment/internal/delivery/rest"
	"testDeployment/internal/usecase"
	"testDeployment/pkg/Bot"
	ai2 "testDeployment/pkg/ai"

	"github.com/gin-gonic/gin"
)

func SetUp(
	g *gin.Engine,
	uc usecase.IUseCase,
	bot Bot.Bot,
	request request.CustomJSONRequester,
	model *ai2.Dermato,
) {
	SetUpHandlerV1(
		g.Group("/api/v1"),
		uc,
		bot,
		request,
		model,
	)

}
func SetUpHandlerV1(
	group *gin.RouterGroup,
	uc usecase.IUseCase,
	bot Bot.Bot,
	request request.CustomJSONRequester,
	model *ai2.Dermato,
) {
	rest.NewFrontend(
		group,
	)
	rest.NewController(
		group,
		uc.IOtherUseCase(),
		bot,
		request,
	)
	rest.NewNewsController(
		group,
		bot,
		uc.INewsUsecase(),
	)
	rest.NewDoctorController(
		group,
		uc.IDoctorUseCase(),
		bot,
	)

	rest.NewFactsController(
		group,
		uc.IFactUseCase(),
	)
	rest.NewChat(
		group,
		model,
	)
}
