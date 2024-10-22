package main

import (
	"encoding/json"
	"testing"
)

func TestCompletion(t *testing.T) {
	ai := newAiClassifier(nil)

	result, err := ai.Classify("This is a test")
	if err == nil {
		t.Errorf("Client is nil, expected error")
	}
	if result != nil {
		t.Errorf("Should not have a result")
	}
}

func TestRemoveMarkdown(t *testing.T) {
	text := "```json\n{\"spamRating\": 0.1, \"summary\": \"This is a test\"}\n```"
	result := removeMarkdown(text)
	if result != "\n{\"spamRating\": 0.1, \"summary\": \"This is a test\"}\n" {
		t.Errorf("Expected markdown to be removed")
	}
	data := &ClassificationResult{
		SpamRating: 0,
		Summary:    "",
	}
	err := json.Unmarshal([]byte(result), data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if data.SpamRating != 0.1 {
		t.Errorf("Expected spam rating to be 0.1, got %f", data.SpamRating)
	}
}
