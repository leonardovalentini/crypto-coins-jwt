package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	portPtr := flag.String("port", "8080", "Port to check the health of the service")
	flag.Parse()

	port := ""

	if portPtr != nil && *portPtr != "" {
		port = *portPtr
	} else {
		port = os.Getenv("SERVER_PORT")
	}

	if port == "" {
		port = "8080"
	}
	url := fmt.Sprintf("http://localhost:%s/health", port)

	// Query the local server
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Exit code 1 signals "Unhealthy" to Docker
		os.Exit(1)
	}
	// Exit code 0 signals "Healthy" to Docker
	os.Exit(0)
}
