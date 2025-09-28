package main

import (
	"log"
	"os"

	"github.com/beleganjur/terraform-provider-tart/tart"
)

func main() {
	addr := os.Getenv("TART_API_ADDR")
	if addr == "" {
		addr = ":8085"
	}
	log.Printf("Starting Tart API controller on %s", addr)
	if err := tart.StartAPIServer(addr); err != nil {
		log.Fatal(err)
	}
}
