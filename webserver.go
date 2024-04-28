package main

import (
	"crypto/md5"
	"fmt"
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

func (h *hash) mailHandler(w http.ResponseWriter, r *http.Request) {

	data, err := readFile(toRelativeUrl(r.URL.Path))
	if err != nil {
		log.Printf("Error reading %s: %v", r.URL.Path, err)
		send404(w)
	}
	parts := strings.Split(r.URL.Path, "/")
	l := len(parts)
	chatId := parts[l-2]
	fn := strings.Replace(parts[l-1], ".html", "", -1)
	hash := r.URL.Query().Get("hash")
	if hash != h.createSimpleHash(chatId+fn) {
		send401(w)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Add("Content-Security-Policy", "default-src 'self'; img-src *; media-src *")
	w.Write(data)
}

func send404(w http.ResponseWriter) {

	data, err := readFile("404.html")
	if err != nil {
		log.Printf("Error reading 404.html: %v", err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func send401(w http.ResponseWriter) {
	data, err := readFile("401.html")
	if err != nil {
		log.Printf("Error reading 401.html: %v", err)
		http.Error(w, "Permissing denied", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)

}

type hash struct {
	salt string
}

func (h *hash) createSimpleHash(key string) string {
	md5 := md5.New()
	md5.Write([]byte(fmt.Sprintf("%s%s", key, h.salt)))
	return fmt.Sprintf("%x", md5.Sum(nil))
}

func WebServer(h *hash) {

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/mail/", h.mailHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
