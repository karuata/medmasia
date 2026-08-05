package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"medmasia/backend/internal/config"
	"medmasia/backend/internal/httpapi"
	"medmasia/backend/internal/service"
	"medmasia/backend/internal/store"
)

func main() {
	settings := config.Load()
	if err := settings.Validate(); err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), settings.RequestTimeout)
	defer cancel()
	st, err := store.Open(ctx, settings.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	agents := service.NewAgentRunner(st, settings)
	server := &http.Server{
		Addr:              settings.Addr,
		Handler:           httpapi.New(settings, st, agents),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("medmasia-sales listening on http://%s", settings.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
