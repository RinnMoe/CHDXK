package main

import (
	"context"
	"flag"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"jcourse/pkg/logx"
)

func main() {
	logx.ConfigureDefault()
	ctx := context.Background()

	sourceDSN := flag.String("source-dsn", "", "PostgreSQL DSN for the source v1 Django database")
	targetDSN := flag.String("target-dsn", "", "PostgreSQL DSN for the target v2 database")
	stateFile := flag.String("state-file", "data/migrate_v1.checkpoint.json", "Checkpoint file for resuming migration")
	flag.Parse()

	if *sourceDSN == "" {
		logx.Fatal(ctx, "--source-dsn flag is required")
	}
	if *targetDSN == "" {
		logx.Fatal(ctx, "--target-dsn flag is required")
	}

	sourceDB, err := gorm.Open(postgres.Open(*sourceDSN), &gorm.Config{})
	if err != nil {
		logx.Fatal(ctx, "connect source database", "err", err)
	}
	targetDB, err := gorm.Open(postgres.Open(*targetDSN), &gorm.Config{})
	if err != nil {
		logx.Fatal(ctx, "connect target database", "err", err)
	}

	migrator := NewMigrator(sourceDB, targetDB, *stateFile)
	if err := migrator.Run(ctx); err != nil {
		logx.Fatal(ctx, "migrate v1 data", "err", err)
	}
}
