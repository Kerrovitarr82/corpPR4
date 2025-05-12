package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// Присоединение
	resp, _ := http.Post("http://localhost:8080/join", "application/json", nil)
	body, _ := ioutil.ReadAll(resp.Body)
	var joinResp map[string]string
	json.Unmarshal(body, &joinResp)
	playerID := joinResp["player_id"]
	fmt.Println("Player ID:", playerID)

	// Попытка угадать
	for {
		var guess string
		fmt.Print("Enter guess: ")
		fmt.Scanln(&guess)

		payload := map[string]string{"player_id": playerID, "guess": guess}
		data, _ := json.Marshal(payload)
		resp, _ := http.Post("http://localhost:8080/guess", "application/json", bytes.NewBuffer(data))
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Server:", string(body))
	}
}
