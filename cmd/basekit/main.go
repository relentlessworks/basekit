package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/relentlessworks/basekit/internal/api"
)

func main() {
	addr := flag.String("addr", "", "listen address (default :7700)")
	flag.Parse()

	// Config: defaults < env < flags
	listenAddr := *addr
	if listenAddr == "" {
		listenAddr = os.Getenv("BASEKIT_ADDR")
	}
	if listenAddr == "" {
		listenAddr = ":7700"
	}

	server := api.NewServer(listenAddr)
	fmt.Fprintf(os.Stderr, "basekit listening on %s\n", listenAddr)
	log.Fatal(server.ListenAndServe())
}
