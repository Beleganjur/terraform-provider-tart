package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/execute", handleExecute)
	log.Println("Starting Executor Daemon at :9090...")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
