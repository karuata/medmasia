package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"medmasia/backend/internal/config"
	"medmasia/backend/internal/service"
	"medmasia/backend/internal/store"
)

func main() {
	path := flag.String("file", "", "private .xlsx file to import")
	maxRows := flag.Int("max-rows", 0, "optional import cap for validation runs")
	flag.Parse()
	if *path == "" {
		fmt.Fprintln(os.Stderr, "-file is required")
		os.Exit(2)
	}

	settings := config.Load()
	if err := settings.Validate(); err != nil {
		log.Fatal(err)
	}
	if *maxRows == 0 {
		*maxRows = settings.MaxImportRows
	}
	ctx, cancel := context.WithTimeout(context.Background(), settings.RequestTimeout)
	defer cancel()
	st, err := store.Open(ctx, settings.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	result, err := service.ImportXLSX(context.Background(), st, *path, *maxRows)
	if err != nil {
		log.Fatal(err)
	}
	buf, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(buf))
}
