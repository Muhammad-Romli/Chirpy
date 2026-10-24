package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	fmt.Printf("Hello, World!")
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serverMux.HandleFunc("/healthz", readinessHandle)
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
	respWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	respWriter.WriteHeader(http.StatusOK)
	_, err := respWriter.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Printf("fail writing response: %v", err)
	}
}
