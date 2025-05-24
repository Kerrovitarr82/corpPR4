package controllers

import (
	"corpPR4/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GameController struct {
	service *services.GameService
}

func NewGameController(s *services.GameService) *GameController {
	return &GameController{service: s}
}

func (gc *GameController) Join(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	player, err := gc.service.AddPlayer(req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"player_id": player.ID})
}

func (gc *GameController) Ready(c *gin.Context) {
	var req struct {
		PlayerID string `json:"player_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PlayerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_id is required"})
		return
	}

	roundActive, err := gc.service.MarkReady(req.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if roundActive {
		c.JSON(http.StatusOK, gin.H{"status": "marked as ready and game starts!"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "marked as ready"})
	}

}

func (gc *GameController) Guess(c *gin.Context) {
	var req struct {
		PlayerID string `json:"player_id"`
		Guess    string `json:"guess"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Guess) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	black, white, message, err := gc.service.ProcessGuess(req.PlayerID, req.Guess)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"black":   black,
		"white":   white,
		"message": message,
	})
}

func (gc *GameController) Status(c *gin.Context) {
	status := gc.service.GetStatus()
	c.JSON(http.StatusOK, status)
}

func (gc *GameController) TurnOrder(c *gin.Context) {
	status := gc.service.GetTurnOrder()
	c.JSON(http.StatusOK, status)
}

func (gc *GameController) Result(c *gin.Context) {
	result := gc.service.GetResult()
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "game is not finished yet"})
		return
	}
	c.JSON(http.StatusOK, result)
}
