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

func mailHandler(w http.ResponseWriter, r *http.Request) {

	data, err := readFile(toRelativeUrl(r.URL.Path))
	if err != nil {
		log.Printf("Error reading %s: %v", r.URL.Path, err)
		send404(w)
	}
	w.Write(data)
}

func send404(w http.ResponseWriter) {
	data, err := readFile("404.html")
	if err != nil {
		log.Printf("Error reading 404.html: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		send404(w)
		return
	}
	data, err := readFile("index.html")
	if err != nil {
		send404(w)
		return
	}
	w.Write(data)
}

func WebServer() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/mail/", mailHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
