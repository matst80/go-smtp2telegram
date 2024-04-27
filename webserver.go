package main

import (
	"log"
	"net/http"
	"strings"
)

func toRelativeUrl(url string) string {
	if url == "/" {
		return "/index.html"
	}
	return strings.Replace(url, "/", "", 1)
}

func isValidUrl(url string) bool {
	return strings.Contains(url, "mail/")
}

func handler(w http.ResponseWriter, r *http.Request) {
	if !isValidUrl(r.URL.Path) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	data, err := readFile(toRelativeUrl(r.URL.Path))
	if err != nil {
		log.Printf("Error reading %s: %v", r.URL.Path, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func WebServer() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
