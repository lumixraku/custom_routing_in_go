package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	handler := http.NewServeMux()

	handler.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.Replace(r.URL.Path, "/hello/", "", 1)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, fmt.Sprintf("Hello %s\n", name))
	})

	handler.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, "Hello world\n")
	})

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)

		io.WriteString(w, "Not found\n")
	})

	err := http.ListenAndServe(":9000", handler)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
