package rest

import (
	"github.com/gin-contrib/sessions"
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
	router.POST("/createQuestions", handler.CreateQuestions)
	router.GET("/getFacts", handler.GetFact)
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
	s := sessions.Default(ctx)
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
	s.Set("factId", fact.Id)
	s.Save()
	ctx.JSON(
		http.StatusCreated,
		gin.H{"Id": fact.Id})
}

// CreateFactHandler godoc
// @Summary Create a fact question
// @Description Creates a new fact question and returns the created fact questions.
// @ID create-fact-question
// @tags fact
// @Produce json
// @Param fact body []dto.FactQuestions true "List of fact questions to be created"
// @Success 201 {array} dto.FactQuestions
// @Router /fact/createQuestions [post]
func (c facts) CreateQuestions(ctx *gin.Context) {
	s := sessions.Default(ctx)
	questions := []dto.FactQuestions{}
	if err := ctx.ShouldBind(&questions); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	id := s.Get("factId").(int)
	err := c.usecase.CreateQuestion(
		ctx.Request.Context(),
		id,
		&questions,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	s.Delete("factId")
	s.Save()

	ctx.JSON(
		http.StatusCreated,
		gin.H{"message": "successfully created"},
	)
}

// CreateFactHandler godoc
// @Summary Get a fact
// @Description Get a 5 facts
// @ID get-fact
// @tags fact
// @Produce json
// @Success 200 {array} dto.Fact
// @Router /fact/getFact [get]
func (c facts) GetFact(ctx *gin.Context) {
	facts, err := c.usecase.GetFacts(
		ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	ctx.JSON(http.StatusOK, facts)
}
