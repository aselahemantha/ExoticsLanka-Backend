package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations executes SQL files in the specified directory against the database
func RunMigrations(ctx context.Context, db *pgxpool.Pool, migrationDir string) error {
	log.Printf("Starting database migrations from: %s", migrationDir)

	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}

	sort.Strings(sqlFiles) // Ensure ordered execution

	for _, filename := range sqlFiles {
		log.Printf("Applying migration: %s", filename)
		path := filepath.Join(migrationDir, filename)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		if _, err := db.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		log.Printf("Successfully applied migration: %s", filename)
	}

	log.Println("All migrations completed successfully")
	return nil
}
