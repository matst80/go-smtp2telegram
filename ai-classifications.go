package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type aiClassifier struct {
	client *openai.Client
}

type classificationResult struct {
	SpamRating float64 `json:"spamRating"`
	Summary    string  `json:"summary"`
}

func (a *aiClassifier) classify(text string, result *classificationResult) error {

	if a.client == nil {
		return fmt.Errorf("OpenAI client is not initialized")
	}
	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a mail analyzer, summarize and classify content and respond in json format with the keys spamRating (0 to 10) and short summary of the content, include passwords and codes if found.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
		},
	)

	if err != nil {
		//fmt.Printf("ChatCompletion error: %v\n", err)
		return err
	}

	c := strings.ReplaceAll(resp.Choices[0].Message.Content, "```json", "")
	c = strings.ReplaceAll(c, "```", "")

	err = json.Unmarshal([]byte(c), result)
	if err != nil {
		fmt.Printf("Error reading result: %v\nInput: %s", err, c)
		return err
	}

	return nil
}

func newAiClassifier(config *AiClassification) *aiClassifier {
	if config == nil {
		return &aiClassifier{}
	}
	client := &openai.Client{}
	if config.ApiKey != "" {
		client = openai.NewClient(config.ApiKey)
	}
	return &aiClassifier{
		client: client,
	}
}
