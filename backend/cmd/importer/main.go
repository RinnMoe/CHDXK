package main

import (
	"context"
	"flag"
	"fmt"

	"jcourse/config"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/pkg/logx"
)

func main() {
	logx.ConfigureDefault()
	ctx := context.Background()

	semester := flag.String("semester", "", "Semester to import (e.g. 2025-2026-1)")
	configPath := flag.String("config", "config/config.yaml", "Config file path")
	flag.Parse()

	if *semester == "" {
		logx.Fatal(ctx, "--semester flag is required")
	}

	conf, err := config.Load(*configPath)
	if err != nil {
		logx.Fatal(ctx, "load config", "err", err)
	}

	db := persistence.NewPostgres(conf.Postgres)

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
