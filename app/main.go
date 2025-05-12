package main

import (
	"corpPR4/server"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	game := server.NewGame()
	go game.Run()

	r.POST("/join", game.JoinHandler)
	r.POST("/guess", game.GuessHandler)

	r.Run(":8080")
}
