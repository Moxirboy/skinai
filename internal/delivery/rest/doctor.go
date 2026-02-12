package rest

import (
	"testDeployment/internal/usecase"
	"testDeployment/pkg/Bot"

	"github.com/gin-gonic/gin"
)

type doctor struct{
	uc usecase.IDoctorUsecase
	bot Bot.Bot
}
func NewDoctorController(g *gin.RouterGroup,uc usecase.IDoctorUsecase,bot Bot.Bot) {
	handler:=doctor{
		uc:uc,
		bot:bot,
	}
	r:=g.Group("/doc")
	r.GET("/getalldoctors",handler.GetAll)
	r.GET("/getonedoctor",handler.GetOneByID)
}

// GetAllDoctors godoc
// @Summary      Get all doctors
// @Description  Returns all doctors grouped by type
// @Tags         doctors
// @Produce      json
// @Success      200  {array}   domain.DoctorByType
// @Failure      500  {object}  map[string]interface{}
// @Router       /doc/getalldoctors [get]
func (r *doctor) GetAll(c *gin.Context){
	doctors,err:=r.uc.GetAll(c)
	if err!=nil{
		c.JSON(200,gin.H{
			"error":err,
		})
		return
	}
	if doctors==nil{
		c.JSON(200,gin.H{
			"message":"no doctors yet",
		})
		return
	}
	c.JSON(200,doctors)
}

// GetOneDoctorByID godoc
// @Summary      Get doctor by name
// @Description  Returns doctor details filtered by name
// @Tags         doctors
// @Produce      json
// @Param        name  query  string  true  "Doctor name"
// @Success      200  {array}   domain.DoctorWithType
// @Failure      500  {object}  map[string]interface{}
// @Router       /doc/getonedoctor [get]
func (r *doctor) GetOneByID(c *gin.Context){
	name:=c.Query("name")
	
	doctor,err:=r.uc.GetOneByID(c,name)
	if err!=nil{
		c.JSON(200,gin.H{
			"error":err,
		})
		return
	}
	if doctor==nil{
		c.JSON(200,gin.H{
			"message":"no doctor yet",
		})
		return
	}
	c.JSON(200,doctor)
}