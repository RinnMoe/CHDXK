package main

import (
	"flag"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	sourceDSN := flag.String("source-dsn", "", "PostgreSQL DSN for the source v1 Django database")
	targetDSN := flag.String("target-dsn", "", "PostgreSQL DSN for the target v2 database")
	stateFile := flag.String("state-file", "data/migrate_v1.checkpoint.json", "Checkpoint file for resuming migration")
	flag.Parse()

	if *sourceDSN == "" {
		log.Fatal("--source-dsn flag is required")
	}
	if *targetDSN == "" {
		log.Fatal("--target-dsn flag is required")
	}

	sourceDB, err := gorm.Open(postgres.Open(*sourceDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Connect source database: %v", err)
	}
	targetDB, err := gorm.Open(postgres.Open(*targetDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Connect target database: %v", err)
	}

	migrator := NewMigrator(sourceDB, targetDB, *stateFile)
	if err := migrator.Run(); err != nil {
		log.Fatalf("Migrate v1 data: %v", err)
	}
}
