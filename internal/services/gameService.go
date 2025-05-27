package services

import (
	"corpPR4/internal/models"
	"corpPR4/internal/utils"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type GameService struct {
	mu                sync.Mutex
	players           map[string]*models.Player
	readyPlayers      map[string]bool
	turnOrder         []string
	currentTurn       int
	secretCode        [2]int
	attempts          map[string]int
	roundActive       bool
	roundStart        time.Time
	roundEnd          time.Time
	result            *models.GameResult
	inactivityTimeout time.Duration
}

func NewGameService() *GameService {
	gs := &GameService{
		players:           make(map[string]*models.Player),
		readyPlayers:      make(map[string]bool),
		attempts:          make(map[string]int),
		inactivityTimeout: 30 * time.Second,
	}

	go gs.monitorInactivity()

	return gs
}

func (gs *GameService) monitorInactivity() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		gs.mu.Lock()

		if !gs.roundActive || len(gs.turnOrder) == 0 {
			gs.mu.Unlock()
			continue
		}

		currentPlayerID := gs.turnOrder[gs.currentTurn]
		player, ok := gs.players[currentPlayerID]
		if !ok {
			gs.advanceTurn()
			gs.mu.Unlock()
			continue
		}

		if time.Since(player.LastActive) > gs.inactivityTimeout {
			fmt.Printf("Player %s AFK, kick him.\n", gs.players[currentPlayerID].Name)
			delete(gs.players, currentPlayerID)
			delete(gs.readyPlayers, currentPlayerID)
			delete(gs.attempts, currentPlayerID)
			gs.turnOrder = append(gs.turnOrder[:gs.currentTurn], gs.turnOrder[gs.currentTurn+1:]...)

			if len(gs.turnOrder) < 2 {
				fmt.Println("Too few players. You need at least 2 players to play.")
				gs.roundActive = false
				gs.readyPlayers = make(map[string]bool)
				gs.turnOrder = nil
				gs.currentTurn = 0
				gs.mu.Unlock()
				continue
			}

			if gs.currentTurn >= len(gs.turnOrder) {
				gs.currentTurn = 0
			}
		} else {
		}

		gs.mu.Unlock()
	}
}

func (gs *GameService) advanceTurn() {
	if len(gs.turnOrder) == 0 {
		return
	}
	gs.currentTurn = (gs.currentTurn + 1) % len(gs.turnOrder)
	currentPlayerID := gs.turnOrder[gs.currentTurn]
	player := gs.players[currentPlayerID]
	player.LastActive = time.Now()
}

func (gs *GameService) AddPlayer(name string) (*models.Player, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.roundActive {
		return nil, errors.New("game in progress; cannot join")
	}

	if len(gs.players) >= 4 {
		return nil, errors.New("maximum 4 players allowed")
	}

	for _, player := range gs.players {
		if player.Name == name {
			return nil, errors.New("player with such name already exists")
		}
	}

	player := models.NewPlayer(name)
	gs.players[player.ID] = player
	return player, nil
}

func (gs *GameService) MarkReady(playerID string) (bool, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	player, ok := gs.players[playerID]
	if !ok {
		return false, errors.New("player not found")
	}
	if gs.readyPlayers[playerID] {
		return false, errors.New("player already marked as ready")
	}

	player.LastActive = time.Now()
	gs.readyPlayers[playerID] = true

	if len(gs.readyPlayers) >= 2 && len(gs.readyPlayers) == len(gs.players) {
		gs.startNewRound()
	}
	return gs.roundActive, nil
}

func (gs *GameService) startNewRound() {
	gs.secretCode = [2]int{rand.Intn(10), rand.Intn(10)}
	log.Println(gs.secretCode)
	gs.turnOrder = make([]string, 0, len(gs.readyPlayers))
	for pid := range gs.readyPlayers {
		gs.turnOrder = append(gs.turnOrder, pid)
	}
	rand.Shuffle(len(gs.turnOrder), func(i, j int) {
		gs.turnOrder[i], gs.turnOrder[j] = gs.turnOrder[j], gs.turnOrder[i]
	})
	gs.attempts = make(map[string]int)
	gs.currentTurn = 0
	gs.roundActive = true
	gs.roundStart = time.Now()
	gs.result = nil
}

func (gs *GameService) ProcessGuess(playerID, guess string) (int, int, string, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if !gs.roundActive {
		return 0, 0, "", errors.New("no active round")
	}

	if len(gs.turnOrder) == 0 {
		return 0, 0, "", errors.New("no players in turn order")
	}

	if gs.turnOrder[gs.currentTurn] != playerID {
		return 0, 0, "", errors.New("not your turn")
	}

	if player, ok := gs.players[playerID]; ok {
		player.LastActive = time.Now()
	}

	guessDigits := [2]int{}
	for i := 0; i < 2; i++ {
		if guess[i] < '0' || guess[i] > '9' {
			return 0, 0, "", errors.New("guess must be digits 0-9")
		}
		guessDigits[i] = int(guess[i] - '0')
	}

	black, white := evaluateGuess(gs.secretCode, guessDigits)
	gs.attempts[playerID]++

	message := "next player's turn"

	if black == 2 {
		gs.roundActive = false
		gs.roundEnd = time.Now()
		var attemptItems []models.AttemptItem
		for pid, count := range gs.attempts {
			attemptItems = append(attemptItems, models.AttemptItem{
				PlayerName: gs.players[pid].Name,
				Count:      count,
			})
		}
		gs.result = &models.GameResult{
			Start:      gs.roundStart,
			End:        gs.roundEnd,
			Code:       gs.secretCode,
			Attempts:   attemptItems,
			WinnerID:   playerID,
			WinnerName: gs.players[playerID].Name,
		}
		go gs.persistResult()
		gs.readyPlayers = make(map[string]bool)
		gs.turnOrder = nil
		message = fmt.Sprintf("Player %s guessed the code!", gs.players[playerID].Name)
	}

	if gs.roundActive {
		gs.advanceTurn()
	}

	return black, white, message, nil
}

func evaluateGuess(code, guess [2]int) (int, int) {
	black, white := 0, 0
	usedCode := [2]bool{}
	usedGuess := [2]bool{}

	for i := 0; i < 2; i++ {
		if guess[i] == code[i] {
			black++
			usedCode[i], usedGuess[i] = true, true
		}
	}
	for i := 0; i < 2; i++ {
		if usedGuess[i] {
			continue
		}
		for j := 0; j < 2; j++ {
			if !usedCode[j] && guess[i] == code[j] {
				white++
				usedCode[j] = true
				break
			}
		}
	}
	return black, white
}

func (gs *GameService) GetStatus() map[string]interface{} {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	return map[string]interface{}{
		"round_active": gs.roundActive,
		"players":      len(gs.players),
		"ready":        len(gs.readyPlayers),
		"current_turn": func() string {
			if gs.roundActive && len(gs.turnOrder) > 0 {
				return gs.players[gs.turnOrder[gs.currentTurn]].Name
			}
			return ""
		}(),
	}
}

func (gs *GameService) GetTurnOrder() []string {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	var turnsNamed []string
	for _, turn := range gs.turnOrder {
		turnsNamed = append(turnsNamed, gs.players[turn].Name)
	}
	return turnsNamed
}

func (gs *GameService) GetResult() *models.GameResult {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	return gs.result
}

func (gs *GameService) persistResult() {
	err := utils.SaveGameResultToXML(gs.result)
	if err != nil {
		fmt.Printf("Failed to save game result: %v\n", err)
	}
}
