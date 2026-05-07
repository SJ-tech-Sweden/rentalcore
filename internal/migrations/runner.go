package migrations

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func isForwardMigrationFile(name string) bool {
	if !strings.HasSuffix(name, ".sql") {
		return false
	}
	lower := strings.ToLower(name)
	if strings.Contains(lower, ".down.") || strings.HasSuffix(lower, ".down.sql") {
		return false
	}
	return true
}

// ApplyMigrations applies all .sql files in dir in lexical order and records
// applied migrations in schema_migrations table.
func ApplyMigrations(db *sql.DB, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && isForwardMigrationFile(f.Name()) {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return err
	}

	for _, name := range sqlFiles {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)", name).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			log.Printf("skipping already applied migration %s", name)
			continue
		}
		path := filepath.Join(dir, name)
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(b)); err != nil {
			return err
		}
		if _, err := db.Exec("INSERT INTO schema_migrations (name) VALUES ($1)", name); err != nil {
			return err
		}
		log.Printf("applied migration %s", name)
	}
	return nil
}

// ApplySeeds applies all .sql files in a seeds directory (lexical order).
// If the directory does not exist, it's a no-op.
func ApplySeeds(db *sql.DB, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && isForwardMigrationFile(f.Name()) {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)
	for _, name := range sqlFiles {
		path := filepath.Join(dir, name)
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(b)); err != nil {
			return err
		}
		log.Printf("applied seed %s", name)
	}
	return nil
}
