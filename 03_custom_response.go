package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Response struct {
	http.ResponseWriter
}

func (r *Response) Text(code int, body string) {
	r.Header().Set("Content-Type", "text/plain")
	r.WriteHeader(code)

	io.WriteString(r, fmt.Sprintf("%s\n", body))
}

func main() {
	handler := http.NewServeMux()

	handler.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.Replace(r.URL.Path, "/hello/", "", 1)

		resp := Response{w}
		resp.Text(http.StatusOK, fmt.Sprintf("Hello %s", name))
	})

	handler.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		resp := Response{w}
		resp.Text(http.StatusOK, "Hello world")
	})

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := Response{w}
		resp.Text(http.StatusNotFound, "Not found")
	})

	err := http.ListenAndServe(":9000", handler)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}

}
