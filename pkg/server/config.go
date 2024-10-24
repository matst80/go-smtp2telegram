package server

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	Token             string              `json:"token"`
	Domain            string              `json:"domain"`
	DkimSelector      string              `json:"dkimSelector"`
	Listen            string              `json:"listen"`
	OpenAi            AiClassification    `json:"openai"`
	CustomFromMessage []CustomFromMessage `json:"customFromMessage"`
	Users             []User              `json:"users"`
	AllowBlockedIps   bool                `json:"allowBlockedIps"`
	StopWords         []string            `json:"stopWords"`
	BlockedIpUrl      string              `json:"blockedIpUrl"`
	WarningWordsUrl   string              `json:"warningWordsUrl"`
	BaseUrl           string              `json:"baseUrl"`
	HashSalt          string              `json:"hashSalt"`
}

type AiClassification struct {
	ApiKey string `json:"apiKey"`
	Url    string `json:"url"`
	Model  string `json:"model"`
}

type CustomFromMessage struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

type User struct {
	Email          string `json:"email"`
	DebugInfo      bool   `json:"debugInfo"`
	DefaultSubject string `json:"defaultSubject"`
	ChatId         int64  `json:"chatId"`
}

func readFile(file string) ([]byte, error) {
	configFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()
	return io.ReadAll(configFile)
}

func parseConfig(bytes []byte) (*Config, error) {
	config := &Config{
		Users:             []User{},
		StopWords:         []string{},
		CustomFromMessage: []CustomFromMessage{},
		Listen:            "0.0.0.0:25",
		HashSalt:          "salty-change-me",
		BaseUrl:           "http://localhost:8080",
	}
	if err := json.Unmarshal([]byte(bytes), config); err != nil {
		return &Config{}, fmt.Errorf("error parsing config: %w", err)
	}
	return config, nil
}

func GetConfig(file string) (*Config, error) {
	bytes, err := readFile(file)
	if err != nil {
		return &Config{}, fmt.Errorf("error reading %s: %w", file, err)
	}
	return parseConfig(bytes)
}
