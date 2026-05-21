package main

import (
	"flag"
	"fmt"
	"log"

	"jcourse/config"
	"jcourse/internal/infrastructure/persistence"
)

func main() {
	semester := flag.String("semester", "", "Semester to import (e.g. 2025-2026-1)")
	configPath := flag.String("config", "config/config.yaml", "Config file path")
	flag.Parse()

	if *semester == "" {
		log.Fatal("--semester flag is required")
	}

	conf, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Load config: %v", err)
	}

	db := persistence.NewPostgres(conf.Postgres)

	filePath := fmt.Sprintf("data/%s.csv", *semester)
	log.Printf("Importing %s from %s ...", *semester, filePath)

	rows, err := parseCSV(filePath)
	if err != nil {
		log.Fatalf("Parse CSV: %v", err)
	}
	log.Printf("Parsed %d rows", len(rows))

	importer := NewImporter(db, *semester)
	if err := importer.Run(rows); err != nil {
		log.Fatalf("Import: %v", err)
	}
}
