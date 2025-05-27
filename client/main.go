package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const serverURL = "http://localhost:8080"

var playerID string

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Добро пожаловать в консольный клиент игры!")
	for {
		fmt.Println("\nВыберите действие:")
		fmt.Println("1. Присоединиться к игре")
		fmt.Println("2. Отметиться как готов")
		fmt.Println("3. Сделать попытку (guess)")
		fmt.Println("4. Посмотреть статус игры")
		fmt.Println("5. Посмотреть результат")
		fmt.Println("6. Посмотреть очередность ходов")
		fmt.Println("0. Выход")
		fmt.Print("> ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			if !isPlayerIDSet() {
				fmt.Print("Введите имя игрока: ")
				name, _ := reader.ReadString('\n')
				name = strings.TrimSpace(name)
				playerID = joinPlayer(name)
			} else {
				fmt.Println("Вы уже зарегистрированы!")
			}
		case "2":
			if isPlayerIDSet() {
				markReady(playerID)
			} else {
				fmt.Println("Сначала присоединитесь к игре!")
			}

		case "3":
			if isPlayerIDSet() {
				fmt.Print("Введите вашу попытку (две цифры): ")
				guess, _ := reader.ReadString('\n')
				guess = strings.TrimSpace(guess)
				makeGuess(playerID, guess)
			} else {
				fmt.Println("Сначала присоединитесь к игре!")
			}
		case "4":
			getStatus()
		case "5":
			getResult()
		case "6":
			getTurnOrder()
		case "0":
			fmt.Println("Выход.")
			return
		default:
			fmt.Println("Неверный ввод.")
		}
	}
}

func isPlayerIDSet() bool {
	if playerID == "" {
		return false
	}
	return true
}

func joinPlayer(name string) string {
	payload := map[string]string{"name": name}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(serverURL+"/join", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Ошибка подключения:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Ответ сервера:", string(body))
		return ""
	}

	var body struct {
		PlayerID string `json:"player_id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	fmt.Println("Вы зарегистрированы с ID:", body.PlayerID)
	return body.PlayerID
}

func markReady(id string) {
	payload := map[string]string{"player_id": id}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(serverURL+"/ready", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Ошибка запроса:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Ответ сервера:", string(body))
}

func makeGuess(id, guess string) {
	payload := map[string]string{
		"player_id": id,
		"guess":     guess,
	}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(serverURL+"/guess", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Ошибка запроса:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Ответ сервера:", string(body))
}

func getStatus() {
	resp, err := http.Get(serverURL + "/status")
	if err != nil {
		fmt.Println("Ошибка запроса:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Статус игры:\n", string(body))
}

func getResult() {
	resp, err := http.Get(serverURL + "/result")
	if err != nil {
		fmt.Println("Ошибка запроса:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Результат игры:\n", string(body))
}

func getTurnOrder() {
	resp, err := http.Get(serverURL + "/turns")
	if err != nil {
		fmt.Println("Ошибка запроса:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Очередь ходов:\n", string(body))
}
