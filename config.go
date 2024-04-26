package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type Config struct {
	Token           string   `json:"token"`
	Domain          string   `json:"domain"`
	Listen          string   `json:"listen"`
	Users           []user   `json:"users"`
	StopWords       []string `json:"stopWords"`
	WarningWords    []string `json:"warningWords"`
	BlockedIps      []string `json:"blockedIps"`
	BlockedIpUrl    string   `json:"blockedIpUrl"`
	WarningWordsUrl string   `json:"warningWordsUrl"`
}

func GetConfig() Config {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Error opening config.json")
	}
	defer configFile.Close()
	bytes, err := io.ReadAll(configFile)
	if err != nil {
		log.Fatal("Error reading config.json")
	}
	var config Config
	if err := json.Unmarshal([]byte(bytes), &config); err != nil {
		log.Fatal("Error parsing config.json: ", err)
	}
	return config
}
