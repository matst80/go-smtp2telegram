package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

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
		log.Fatal("Error parsing config.json")
	}
	return config
}
