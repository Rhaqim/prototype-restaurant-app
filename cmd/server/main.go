package main

import (
	"github.com/Rhaqim/thedutchapp/pkg/handlers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
)

func main() {
	run := handlers.GinRouter()
	port := ut.GetEnv("PORT")

	if port == "" {
		port = "8080"
	}

	run.Run(port)
}
