package main

import (
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/", home)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func home(w http.ResponseWriter, _ *http.Request) {
	_, err := io.WriteString(w, "Hello world")
	if err != nil {
		panic(err)
	}
}
