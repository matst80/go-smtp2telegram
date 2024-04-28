package main

import (
	"encoding/json"
	"testing"
)

func TestCompletion(t *testing.T) {
	ai := newAiClassifier(nil)
	result := &classificationResult{
		SpamRating: 0,
		Summary:    "",
	}
	err := ai.classify("This is a test", result)
	if err == nil {
		t.Errorf("Client is nil, expected error")
	}
	if result.Summary != "" {
		t.Errorf("Summary is empty")
	}
	if result.SpamRating != 0 {
		t.Errorf("SpamRating has been set")
	}
}

func TestRemoveMarkdown(t *testing.T) {
	text := "```json\n{\"spamRating\": 0.1, \"summary\": \"This is a test\"}\n```"
	result := removeMarkdown(text)
	if result != "\n{\"spamRating\": 0.1, \"summary\": \"This is a test\"}\n" {
		t.Errorf("Expected markdown to be removed")
	}
	data := &classificationResult{
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
