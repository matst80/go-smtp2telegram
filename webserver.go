package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func toRelativeUrl(url string) string {
	if url == "/" {
		return "/index.html"
	}
	return strings.Replace(url, "/", "", 1)
}

func (h *hash) mailHandler(w http.ResponseWriter, r *http.Request) {

	data, err := readFile(toRelativeUrl(r.URL.Path))
	if err != nil {
		log.Printf("Error reading %s: %v", r.URL.Path, err)
		send404(w)
	}
	parts := strings.Split(r.URL.Path, "/")
	chatId, parseError := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	if parseError != nil {
		send404(w)
		return
	}
	hash := r.URL.Query().Get("hash")
	if hash != h.createSimpleHash(chatId) {
		send404(w)
		return
	}
	w.Header().Set("Content-Type", "text/html")
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

type hash struct {
	salt string
}

func (h *hash) createSimpleHash(key int64) string {
	md5 := md5.New()
	md5.Write([]byte(fmt.Sprintf("%d%s", key, h.salt)))
	return fmt.Sprintf("%x", md5.Sum(nil))
}

func WebServer(h *hash) {

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/mail/", h.mailHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
