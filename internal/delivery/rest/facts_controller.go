package rest

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	router.GET("/get-question", handler.GetQuestion)
	router.POST("/answer-question", handler.AnswerQuestion)
	router.POST("/upload", handler.upload)
	router.GET("/get-image", handler.GetImage)
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
// @Router /fact/get-question [get]
func (c facts) GetQuestion(ctx *gin.Context) {
	// Retrieve the 'id' and 'offset' query parameters
	idStr := ctx.Query("id")
	offsetStr := ctx.Query("offset")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		id = 0 // Default value if conversion fails
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0 // Default value if conversion fails
	}

	facts, err := c.usecase.GetQuestion(
		ctx.Request.Context(),
		id,
		offset,
	) // Respond with the extracted parameters
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	ctx.JSON(http.StatusOK, facts)
}

// @Summary Answer a question and update points
// @Description Receives a score and updates the user's points if the score is above a certain threshold
// @Tags fact
// @Accept json
// @Produce json
// @Param score body dto.Score true "Score details"
// @Success 200
// @Router /fact/answer-question [post]
func (c facts) AnswerQuestion(ctx *gin.Context) {
	s := sessions.Default(ctx)
	score := dto.Score{}
	id := s.Get("userId").(int)
	if err := ctx.ShouldBind(&score); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	if score.Score == 0 {
		ctx.JSON(http.StatusOK, gin.H{"message": "score is below  80 No points"})
		return
	}
	if (float64(score.Score)/float64(score.NumberOfQuestion))*100 <= 80 {
		ctx.JSON(http.StatusOK, gin.H{"message": "score is below  80 No points"})
		return
	}
	bonus, err := c.usecase.UpdatePoint(
		ctx.Request.Context(),
		id,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(
		http.StatusOK, gin.H{
			"score": bonus,
		},
	)
}

// @Summary Upload an image
// @Description Uploads an image with an ID
// @Tags image
// @Accept multipart/form-data
// @Produce json
// @Param id formData string true "Image ID"
// @Param image formData file true "Image file"
// @Success 200 {string} string "image"
// @Failure 400
// @Failure 500
// @Router /fact/upload [post]
func (f facts) upload(c *gin.Context) {
	// Get the ID from the form data
	id := c.PostForm("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	// Parse the multipart form to get the image
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image from request"})
		return
	}
	defer file.Close()

	filename := id + filepath.Ext(header.Filename)
	filePath := filepath.Join("uploads", filename)

	// Ensure the uploads directory exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", os.ModePerm)
	}

	// Write the file to the local filesystem
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save file"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
		return
	}
	ID, err := strconv.Atoi(id)
	err = f.usecase.UpdateImage(c.Request.Context(), ID, "https://web.binaryhood.uz/api/v1/fact/get-image/?filepath="+filePath)
	// Return the ID, filename, and URL as JSON
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to saving file"})
		return
	}
	c.JSON(http.StatusOK, "image")
}

// @Summary Get an image
// @Description Retrieves an image by its file path
// @Tags image
// @Produce json
// @Param filepath query string true "File path"
// @Success 200 {file} file
// @Failure 400 message invalid
// @Failure 404 message not found file
// @Router /fact/get-image/ [get]
func (f facts) GetImage(c *gin.Context) {
	filePath := c.Query("filepath")

	// Sanitize the file path to prevent directory traversal attacks
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path"})
		return
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Serve the file
	c.File(filePath)
}
