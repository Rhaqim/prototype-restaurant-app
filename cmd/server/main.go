package main

import (
	"os"

	"github.com/Rhaqim/thedutchapp/pkg/handlers"
)

func main() {
	run := handlers.GinRouter()
	port := os.Getenv("PORT")

	run.Run(port)
}
