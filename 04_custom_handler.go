package main

import (
	"io"
	"log"
	"net/http"
)

type App struct{}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, "Hello world\n")
}

func main() {
	err := http.ListenAndServe(":9000", &App{})

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
