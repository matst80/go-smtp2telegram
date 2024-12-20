package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type webserver struct {
	hashGenerator HashGenerator
}

func toRelativeUrl(url string) string {
	if url == "/" {
		return "/index.html"
	}
	return strings.Replace(url, "/", "", 1)
}

func (h *webserver) validateHash(url *url.URL) (bool, string) {
	hash := url.Query().Get("hash")
	parts := strings.Split(url.Path, "/")
	l := len(parts)
	chatId := parts[l-2]
	fn := parts[l-1]
	valid := h.hashGenerator.CreateHash(chatId+fn) == hash
	if !valid && strings.HasSuffix(fn, ".html") {
		fn = strings.TrimSuffix(fn, ".html")
		valid = h.hashGenerator.CreateHash(chatId+fn) == hash
	}
	return valid, fn
}

func (h *webserver) mailHandler(w http.ResponseWriter, r *http.Request) {

	data, err := readFile(toRelativeUrl(r.URL.Path))
	if err != nil {
		log.Printf("Error reading %s: %v", r.URL.Path, err)
		send404(w)
		return
	}
	valid, fileName := h.validateHash(r.URL)
	if !valid {
		send401(w)
		return
	}
	if strings.HasSuffix(r.URL.Path, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	}

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

func WebServer(h *SimpleHash) {
	webserver := &webserver{hashGenerator: h}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("GET /mail/", webserver.mailHandler)
	//http.HandleFunc("POST", "/send-mail", h.sendMail)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
