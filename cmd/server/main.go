package main

import (
	"os"

	"github.com/Rhaqim/thedutchapp/pkg/handlers"
)

func main() {
	run := handlers.GinRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	run.Run(port)
}
