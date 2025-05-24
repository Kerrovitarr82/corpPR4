package models

import "time"

type GameResult struct {
	Start      time.Time     `xml:"start"`
	End        time.Time     `xml:"end"`
	Code       [2]int        `xml:"code"`
	Attempts   []AttemptItem `xml:"attempts>attempt"`
	WinnerID   string        `xml:"winner_id"`
	WinnerName string        `xml:"winner_name"`
}

type AttemptItem struct {
	PlayerName string `xml:"player_name,attr"`
	Count      int    `xml:"count"`
}
