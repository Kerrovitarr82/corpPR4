package server

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Player struct {
	ID       string
	Attempts int
}

type Game struct {
	sync.Mutex
	cond       *sync.Cond
	SecretCode string
	Players    map[string]*Player
	MaxPlayers int
	MaxTries   int
	Started    bool
	Winner     string
	StartTime  time.Time
}

type Result struct {
	XMLName  xml.Name       `xml:"GameResult"`
	Start    time.Time      `xml:"StartTime"`
	End      time.Time      `xml:"EndTime"`
	Code     string         `xml:"SecretCode"`
	Attempts []PlayerResult `xml:"Attempts>Player"`
	Winner   string         `xml:"Winner"`
}

type PlayerResult struct {
	ID    string `xml:"ID,attr"`
	Tries int    `xml:"Tries"`
}

func NewGame() *Game {
	g := &Game{
		SecretCode: generateCode(1),
		Players:    make(map[string]*Player),
		MaxPlayers: 4,
		MaxTries:   10,
	}
	g.cond = sync.NewCond(&g.Mutex)
	return g
}

func (g *Game) Run() {
	for {
		g.Lock()
		for len(g.Players) < 2 {
			fmt.Println("Waiting for at least 2 players...")
			g.cond.Wait()
		}
		g.Started = true
		g.StartTime = time.Now()
		fmt.Println("Game started with players:", len(g.Players))
		g.Unlock()

		// Ждём завершения раунда (проверяем Winner)
		for {
			time.Sleep(1 * time.Second)
			g.Lock()
			if g.Winner != "" {
				g.saveResult()
				g.reset()
				g.Unlock()
				break
			}
			g.Unlock()
		}
	}
}

func generateCode(length int) string {
	chars := "0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}

func (g *Game) JoinHandler(c *gin.Context) {
	g.Lock()
	defer g.Unlock()

	if len(g.Players) >= g.MaxPlayers {
		c.JSON(400, gin.H{"error": "game full"})
		return
	}

	playerID := fmt.Sprintf("player%d", len(g.Players)+1)
	g.Players[playerID] = &Player{ID: playerID}

	// Пробуждаем Run() если >= 2 игроков
	if len(g.Players) >= 2 {
		g.cond.Signal()
	}

	c.JSON(200, gin.H{"player_id": playerID})
}

func (g *Game) GuessHandler(c *gin.Context) {
	type GuessRequest struct {
		PlayerID string `json:"player_id"`
		Guess    string `json:"guess"`
	}

	var req GuessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	g.Lock()
	defer g.Unlock()

	player, exists := g.Players[req.PlayerID]
	if !exists {
		c.JSON(400, gin.H{"error": "player not found"})
		return
	}

	// 🚫 Игра ещё не началась
	if !g.Started {
		c.JSON(200, gin.H{"message": "waiting for other players"})
		return
	}

	// ✅ Игра уже завершена
	if g.Winner != "" {
		c.JSON(200, gin.H{"message": "round over", "winner": g.Winner})
		return
	}

	// 🚫 У игрока закончились попытки
	if player.Attempts >= g.MaxTries {
		c.JSON(200, gin.H{"message": "no attempts left"})
		return
	}

	if len(g.SecretCode) != len(req.Guess) {
		c.JSON(400, gin.H{"message": "wrong length"})
		return
	}

	// Обработка догадки
	player.Attempts++
	guess := strings.ToUpper(req.Guess)
	black, white := CheckGuess(g.SecretCode, guess)

	if black == len(g.SecretCode) {
		g.Winner = req.PlayerID
		c.JSON(200, gin.H{"black": black, "white": white, "message": "correct", "winner": req.PlayerID})
		return
	}

	c.JSON(200, gin.H{"black": black, "white": white})
}

func (g *Game) saveResult() {
	result := Result{
		Start:    g.StartTime,
		End:      time.Now(),
		Code:     g.SecretCode,
		Winner:   g.Winner,
		Attempts: []PlayerResult{},
	}

	for _, p := range g.Players {
		result.Attempts = append(result.Attempts, PlayerResult{ID: p.ID, Tries: p.Attempts})
	}

	data, err := xml.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling XML:", err)
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("../results/result_%d.xml", timestamp)
	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
	}
}

func (g *Game) reset() {
	g.SecretCode = generateCode(4)
	g.Players = make(map[string]*Player)
	g.Winner = ""
	g.Started = false
}

func CheckGuess(secret, guess string) (black, white int) {
	usedSecret := make([]bool, len(secret))
	usedGuess := make([]bool, len(guess))

	for i := range secret {
		if guess[i] == secret[i] {
			black++
			usedSecret[i] = true
			usedGuess[i] = true
		}
	}

	for i := range guess {
		if usedGuess[i] {
			continue
		}
		for j := range secret {
			if !usedSecret[j] && guess[i] == secret[j] {
				white++
				usedSecret[j] = true
				break
			}
		}
	}

	return
}
