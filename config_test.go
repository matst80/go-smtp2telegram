package main

import "testing"

func TestExampleConfigRead(t *testing.T) {
	config, err := GetConfig("config.example.json")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	if len(config.Users) == 0 {
		t.Errorf("Expected config to include users")
	}
	if len(config.StopWords) == 0 {
		t.Errorf("Expected config to include stop words")
	}
	if config.Token == "" {
		t.Errorf("Expected config to include token")
	}
	if config.BaseUrl == "" {
		t.Errorf("Expected config to include a baseUrl")
	}
	if config.Domain == "" {
		t.Errorf("Expected config to include domain")
	}
	if config.Listen == "" {
		t.Errorf("Expected config to include listen")
	}
}

func TestConfigDefaults(t *testing.T) {
	config, err := parseConfig([]byte(`{}`))
	if err != nil {
		t.Errorf("Expected no parsing error, got %v", err)
	}
	if config.Listen != "0.0.0.0:25" {
		t.Errorf("Expected default listen to be set, got %s", config.Listen)
	}
	if config.Users == nil {
		t.Errorf("Expected users to be initialized")
	}
}

func TestConfigReadError(t *testing.T) {
	_, err := GetConfig("config.example_non_existing.json")
	if err == nil {
		t.Errorf("Expected error")
	}
	_, err = parseConfig([]byte(`{`))
	if err == nil {
		t.Errorf("Expected error")
	}
}
