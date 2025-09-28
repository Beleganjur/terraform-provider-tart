package tart

import (
	"log"
	"net/http"
)

// Import handlers and middleware as needed

// StartAPIServer runs the API controller HTTP server.
func StartAPIServer(addr string) error {
	r := SetupRouter()
	log.Println("Starting API Controller at " + addr + "...")
	return http.ListenAndServe(addr, r)
}
