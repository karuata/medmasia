package main

import (
	"context"
	"log"
	"os"

	"medmasia/backend/internal/config"
	"medmasia/backend/internal/mcp"
	"medmasia/backend/internal/store"
)

func main() {
	log.SetOutput(os.Stderr)
	settings := config.Load()
	if err := settings.Validate(); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	st, err := store.Open(ctx, settings.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	if err := mcp.New(settings, st, os.Stdin, os.Stdout).Run(ctx); err != nil {
		log.Fatal(err)
	}
}
