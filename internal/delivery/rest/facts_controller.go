package rest

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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
	router.GET("/getFact", handler.GetFact)
	router.GET("/GetQuestion", handler.GetQuestion)
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

// @Summary Get ID and Offset
// @Description Retrieve the ID and offset from the query parameters.
// @Tags fact
// @Accept  json
// @Produce  json
// @Param id query string false "ID" default("default_id")
// @Param offset query string false "Offset" default("0")
// @Success 200 {object} dto.FactQuestions
// @Router /fact/GetQuestion [get]
func (c facts) GetQuestion(ctx *gin.Context) {
	// Retrieve the 'id' and 'offset' query parameters
	idStr := ctx.Query("id")
	offsetStr := ctx.Query("offset")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		id = 0 // Default value if conversion fails
	}
	fmt.Println(id)

	offset, err := strconv.Atoi(offsetStr)
	fmt.Println(offset)
	if err != nil {
		offset = 0 // Default value if conversion fails
	}

	facts, err := c.usecase.GetQuestion(
		ctx.Request.Context(),
		id,
		offset,
	) // Respond with the extracted parameters

	ctx.JSON(http.StatusOK, facts)
}
