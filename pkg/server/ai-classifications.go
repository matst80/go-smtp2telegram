package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type SpamClassification interface {
	Classify(text string) (*ClassificationResult, error)
}

type OpenAiClassifier struct {
	client *openai.Client
}

type ClassificationResult struct {
	SpamRating float64 `json:"spamRating"`
	Summary    string  `json:"summary"`
	Success    bool    `json:"-"`
}

func removeMarkdown(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, "```json", ""), "```", "")
}

const SystemPrompt = "You are a mail analyzer, summarize and classify content and respond in json format with the keys spamRating (0 to 10) and short summary of the content, include passwords and codes if found."

func (a *OpenAiClassifier) Classify(text string) (*ClassificationResult, error) {

	if a.client == nil {
		return nil, fmt.Errorf("OpenAI client is not initialized")
	}
	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: SystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	c := removeMarkdown(resp.Choices[0].Message.Content)
	result := &ClassificationResult{Success: false}
	err = json.Unmarshal([]byte(c), result)
	if err != nil {
		fmt.Printf("Error reading result: %v\nInput: %s", err, c)
		return nil, err
	}
	result.Success = true

	return result, nil
}

func MakeAiClassifier(config *AiClassification) SpamClassification {
	if config == nil {
		return &OpenAiClassifier{}
	}
	client := &openai.Client{}
	if config.ApiKey != "" {
		client = openai.NewClient(config.ApiKey)
	}
	return &OpenAiClassifier{
		client: client,
	}
}
