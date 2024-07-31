package delivery

import (
	request "testDeployment/internal/delivery/http"
	"testDeployment/internal/delivery/rest"
	"testDeployment/internal/usecase"
	"testDeployment/pkg/Bot"

	"github.com/gin-gonic/gin"
)

func SetUp(
	g *gin.Engine,
	uc usecase.IUseCase,
	bot Bot.Bot,
	request request.CustomJSONRequester,
) {
	SetUpHandlerV1(
		g.Group("/api/v1"),
		uc,
		bot,
		request,
	)

}
func SetUpHandlerV1(
	group *gin.RouterGroup,
	uc usecase.IUseCase,
	bot Bot.Bot,
	request request.CustomJSONRequester,
) {
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
	rest.NewSchedule(
		group,
		uc.IScheduleUseCase(),
		bot,
		request,
	)
	rest.NewFactsController(
		group,
		uc.IFactUseCase(),
	)
}
