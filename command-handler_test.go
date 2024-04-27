package main

import "testing"

type testStruct struct {
	Stuff string `json:"stuff"`
	Nr    int64  `json:"nr"`
}

func TestGetMessageJson(t *testing.T) {
	ts := testStruct{Stuff: "stuff", Nr: 42}
	expected := "<code><pre>{\n  \"stuff\": \"stuff\",\n  \"nr\": 42\n}</pre></code>"
	if actual := getMessageJson(ts); actual != expected {
		t.Errorf("getMessageJson(%v) = %v, want %v", ts, actual, expected)
	}

	expected = "<i>Error serializing data</i>"
	if actual := getMessageJson(make(chan int)); actual != expected {
		t.Errorf("getMessageJson(%v) = %v, want %v", make(chan int), actual, expected)
	}
}
