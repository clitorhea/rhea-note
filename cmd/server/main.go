package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/clitorhea/rhea-note/pkg/api"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	storeDir := flag.String("store", "./server-data", "Directory to store encrypted notes")
	token := flag.String("token", "secret-token", "Authentication token")
	flag.Parse()

	server := api.NewServer(*storeDir, *token)
	
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting secnotes sync server on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
