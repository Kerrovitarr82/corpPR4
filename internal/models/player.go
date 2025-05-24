package models

import (
	"github.com/google/uuid"
	"time"
)

type Player struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	LastActive time.Time `json:"-"`
}

func NewPlayer(name string) *Player {
	return &Player{
		ID:         uuid.New().String(),
		Name:       name,
		LastActive: time.Now(),
	}
}
