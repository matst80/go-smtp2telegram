package main

import "testing"

// test ip blocking
func TestAllowedAddress(t *testing.T) {
	err := AllowedAddress("1.1.1.1")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestBlockedAddress(t *testing.T) {
	err := AllowedAddress("45.88.90.75")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
