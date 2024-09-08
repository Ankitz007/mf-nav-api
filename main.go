// This is only meant for local development
package main

import (
	"fmt"
	"net/http"

	handler "github.com/Ankitz007/mf-nav-api/api"
)

func main() {
	// Register the Handler function to the default router
	http.HandleFunc("/", handler.Handler)

	// Start the HTTP server
	// Note: ":8080" is the port number; you can choose a different one if needed.
	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
