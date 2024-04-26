package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	Token           string   `json:"token"`
	Domain          string   `json:"domain"`
	Listen          string   `json:"listen"`
	Users           []User   `json:"users"`
	StopWords       []string `json:"stopWords"`
	BlockedIpUrl    string   `json:"blockedIpUrl"`
	WarningWordsUrl string   `json:"warningWordsUrl"`
}

type User struct {
	Email  string `json:"email"`
	ChatId int64  `json:"chatId"`
}

func readFile(file string) ([]byte, error) {
	configFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()
	return io.ReadAll(configFile)
}

func parseConfig(bytes []byte) (Config, error) {
	config := Config{
		Users:     []User{},
		StopWords: []string{},
		Listen:    "0.0.0.0:25",
	}
	if err := json.Unmarshal([]byte(bytes), &config); err != nil {
		return Config{}, fmt.Errorf("error parsing config: %w", err)
	}
	return config, nil
}

func GetConfig(file string) (Config, error) {
	bytes, err := readFile(file)
	if err != nil {
		return Config{}, fmt.Errorf("error reading %s: %w", file, err)
	}
	return parseConfig(bytes)
}
