package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	fmt.Printf("Hello, World!")
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", readinessHandle)
	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Fail running the server")
	}
}

func readinessHandle(respWriter http.ResponseWriter, req *http.Request) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatalf("fail reading request body: %w", err)
	}
	respWriter.WriteHeader(200)
	_, err = respWriter.Write(data)
	if err != nil {
		log.Fatalf("fail writing data: %w", err)
	}
}
