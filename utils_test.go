package main

import "testing"

func TestEmailAddressSplitting(t *testing.T) {
	emails := getValidEmailAddresses("a@b.se test hej test+1@a.se")
	if len(emails) != 2 {
		t.Errorf("Expected 2 emails, got %v", emails)
	}
	if emails[0] != "a@b.se" {
		t.Errorf("Expected first email to be a@b.se, got %v", emails[0])
	}
	if emails[1] != "test+1@a.se" {
		t.Errorf("Expected second email to be test+1@a.se, got %v", emails[1])
	}
}
