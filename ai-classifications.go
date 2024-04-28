package main

import (
	"context"
	"encoding/json"
	"fmt"

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
					Content: "You are a mail analyzer, summarize and classify content and respond in json format with only the keys spamRating (0 to 10) and short summary",
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

	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), result)
	if err != nil {
		fmt.Printf("Error reading result: %v\nInput: %s", err, resp.Choices[0].Message.Content)
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
