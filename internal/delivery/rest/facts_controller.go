package rest

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testDeployment/internal/delivery/dto"
	"testDeployment/internal/usecase"
)

type facts struct {
	usecase usecase.IFactUseCase
}

func NewFactsController(
	r *gin.RouterGroup,
	uc usecase.IFactUseCase,
) {
	handler := &facts{uc}
	router := r.Group("/fact")
	router.POST("/create", handler.NewFact)
}

// CreateFactHandler godoc
// @Summary create fact
// @Description create fact
// @ID create-fact
// @tags fact
// @Produce json
// @Param user body dto.Fact true "Fact"
// @Success 201 {object} dto.Fact
// @Router /fact/create [post]
func (c facts) NewFact(ctx *gin.Context) {
	fact := dto.Fact{}

	if err := ctx.ShouldBind(&fact); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	err := c.usecase.CreateFact(
		ctx.Request.Context(),
		&fact,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	ctx.JSON(
		http.StatusCreated,
		gin.H{"Id": fact.Id})
}
