package transport

import (
	"corpPR4/internal/controllers"
	"corpPR4/internal/services"
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	gameService := services.NewGameService()
	controller := controllers.NewGameController(gameService)

	r.POST("/join", controller.Join)
	r.POST("/ready", controller.Ready)
	r.POST("/guess", controller.Guess)
	r.GET("/status", controller.Status)
	r.GET("/turns", controller.TurnOrder)
	r.GET("/result", controller.Result)
}
