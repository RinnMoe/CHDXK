package main

import (
	"context"
	"flag"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"jcourse/pkg/logx"
)

func main() {
	logx.ConfigureDefault()
	ctx := context.Background()

	semester := flag.String("semester", "", "Semester to import (e.g. 2025-2026-1)")
	targetDSN := flag.String("target-dsn", "", "PostgreSQL DSN for the target database")
	flag.Parse()

	if *semester == "" {
		logx.Fatal(ctx, "--semester flag is required")
	}
	if *targetDSN == "" {
		logx.Fatal(ctx, "--target-dsn flag is required")
	}

	db, err := gorm.Open(postgres.Open(*targetDSN), &gorm.Config{})
	if err != nil {
		logx.Fatal(ctx, "connect target database", "err", err)
	}

	filePath := fmt.Sprintf("data/%s.csv", *semester)
	logx.Info(ctx, "importing course data", "semester", *semester, "file_path", filePath)

	rows, err := parseCSV(filePath)
	if err != nil {
		logx.Fatal(ctx, "parse csv", "err", err)
	}
	logx.Info(ctx, "parsed csv rows", "count", len(rows))

	importer := NewImporter(db, *semester)
	if err := importer.Run(ctx, rows); err != nil {
		logx.Fatal(ctx, "import", "err", err)
	}
}
