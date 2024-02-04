package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// "steam_api_key" : "75958C8B399D3BABB85451223B5AFAD4",
// "steam_api_format": "json",
// "steam_api_server": "https://api.steampowered.com",
// "d2match_service": "/IDOTA2Match_570",
// "d2match_function": "/GetMatchHistory",
// "d2match_version": "v1",

type Config struct {
	ApiInterval     int    `json:"api_interval"`
	SteamApiFormat  string `json:"steam_api_format"`
	SteamApiKey     string `json:"steam_api_key"`
	SteamApiServer  string `json:"steam_api_server"`
	D2MatchService  string `json:"d2match_service"`
	D2MatchFunction string `json:"d2match_function"`
	D2MatchVersion  string `json:"d2match_version"`

	SteamAccountID string `json:"steam_account_id"`
}

type MatchHistory struct {
	Result struct {
		Status           uint8  `json:"status"`
		NumResults       uint16 `json:"num_results"`
		TotalResults     uint16 `json:"total_results"`
		ResultsRemaining uint16 `json:"results_remaining"`
		Matches          []struct {
			MatchID       uint64 `json:"match_id"`
			MatchSeqNum   uint64 `json:"match_seq_num"`
			StartTime     uint64 `json:"start_time"`
			LobbyType     uint8  `json:"lobby_type"`
			RadiantTeamID uint64 `json:"radiant_team_id"`
			DireTeamID    uint64 `json:"dire_team_id"`
			Players       []struct {
				AccountID  uint64 `json:"account_id"`
				PlayerSlot uint8  `json:"player_slot"`
				TeamNumber uint8  `json:"team_number"`
				TeamSlot   uint8  `json:"team_slot"`
				HeroID     uint8  `json:"hero_id"`
			} `json:"players"`
		} `json:"matches"`
	} `json:"result"`
}

func main() {
	config, err := readConfig("config.json")

	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}
	// Run the service in a Goroutine
	go runService(config)

	// Keep the main Goroutine alive
	// Otherwise, the program will exit immediately
	select {}
}

func runService(config Config) {
	// Define the interval
	interval := time.Duration(config.ApiInterval) * time.Second

	// Create a ticker that ticks every 'interval' duration
	ticker := time.NewTicker(interval)

	// Run the service loop
	for {
		// Wait for the ticker to tick
		<-ticker.C

		previousMatch := getPreviousGame()
		absoluteApiUrl := fmt.Sprintf("https://api.steampowered.com/IDOTA2Match_570/GetMatchHistory/v1/?key=%s&account_id=%s&format=json", config.SteamApiKey, config.SteamAccountID)
		jsonReturn := makeRequestMatchHistory(absoluteApiUrl)

		if previousMatch != jsonReturn.Result.Matches[0].MatchID {
			absoluteApiUrl = fmt.Sprintf("https://api.steampowered.com/IDOTA2Match_570/GetMatchDetails/v1/?key=%s&match_id=%d&format=json", config.SteamApiKey, jsonReturn.Result.Matches[0].MatchID)
			fmt.Println(absoluteApiUrl)
			jsonReturn = makeRequestMatchHistory(absoluteApiUrl)
		}

		// Perform the task
		fmt.Println(jsonReturn.Result)
		fmt.Println("Task performed at", time.Now())

		fmt.Println(jsonReturn)
	}
}

func recordPreviousGame(match_id string) {
	fileName := "PreviousMatch.txt"

	// Open the file for writing
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Data to write to the file
	data := []byte(match_id)

	// Write data to the file
	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("Data written to file successfully!")
}

func getPreviousGame() uint64 {
	// Specify the file path
	fileName := "PreviousMatch.txt"

	// Open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nul
	}
	defer file.Close()

	// Create a buffer to read the file content
	buffer := make([]byte, 1024)

	// Read the file content
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		fmt.Println("Error reading file:", err)
		return ""
	}

	// Print the contents of the file
	fmt.Println("File contents:")
	return uint64(buffer[:n])
}

func makeRequestMatchHistory(url string) MatchHistory {
	response, err := http.Get(url)
	var responseData MatchHistory
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
	} else {
		defer response.Body.Close()

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
		}

		// Unmarshal the JSON response into a struct
		fmt.Println("Response:", body)
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			fmt.Println("Error decoding JSON response:", err)
		} else {
			return responseData
		}
	}
	return responseData
}

func readConfig(filename string) (Config, error) {
	var config Config

	// Read the file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	// Unmarshal JSON into Config struct
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
