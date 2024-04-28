package main

import "testing"

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
