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
		for _, stmt := range splitSQLStatements(string(b)) {
			if _, err := db.Exec(stmt); err != nil {
				return err
			}
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
		for _, stmt := range splitSQLStatements(string(b)) {
			if _, err := db.Exec(stmt); err != nil {
				return err
			}
		}
		log.Printf("applied seed %s", name)
	}
	return nil
}

func splitSQLStatements(sqlText string) []string {
	var (
		statements   []string
		start        int
		inSingle     bool
		inDouble     bool
		inLine       bool
		inBlock      bool
		dollarTag    string
		textLen      = len(sqlText)
	)

	for i := 0; i < textLen; i++ {
		if inLine {
			if sqlText[i] == '\n' {
				inLine = false
			}
			continue
		}
		if inBlock {
			if i+1 < textLen && sqlText[i] == '*' && sqlText[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if dollarTag != "" {
			if strings.HasPrefix(sqlText[i:], dollarTag) {
				i += len(dollarTag) - 1
				dollarTag = ""
			}
			continue
		}
		if inSingle {
			if sqlText[i] == '\'' {
				if i+1 < textLen && sqlText[i+1] == '\'' {
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDouble {
			if sqlText[i] == '"' {
				inDouble = false
			}
			continue
		}

		if i+1 < textLen && sqlText[i] == '-' && sqlText[i+1] == '-' {
			inLine = true
			i++
			continue
		}
		if i+1 < textLen && sqlText[i] == '/' && sqlText[i+1] == '*' {
			inBlock = true
			i++
			continue
		}
		if sqlText[i] == '\'' {
			inSingle = true
			continue
		}
		if sqlText[i] == '"' {
			inDouble = true
			continue
		}
		if sqlText[i] == '$' {
			if tag, ok := matchDollarTag(sqlText, i); ok {
				dollarTag = tag
				i += len(tag) - 1
				continue
			}
		}
		if sqlText[i] == ';' {
			stmt := strings.TrimSpace(sqlText[start:i])
			if isExecutableSQL(stmt) {
				statements = append(statements, stmt)
			}
			start = i + 1
		}
	}

	if tail := strings.TrimSpace(sqlText[start:]); isExecutableSQL(tail) {
		statements = append(statements, tail)
	}
	return statements
}

func matchDollarTag(sqlText string, start int) (string, bool) {
	if start >= len(sqlText) || sqlText[start] != '$' {
		return "", false
	}
	end := start + 1
	for end < len(sqlText) {
		ch := sqlText[end]
		if ch == '$' {
			return sqlText[start : end+1], true
		}
		if !isValidDollarTagChar(ch) {
			return "", false
		}
		end++
	}
	return "", false
}

func isValidDollarTagChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}

func isExecutableSQL(stmt string) bool {
	if stmt == "" {
		return false
	}
	var (
		builder   strings.Builder
		inSingle  bool
		inDouble  bool
		inLine    bool
		inBlock   bool
		dollarTag string
		n         = len(stmt)
	)
	for i := 0; i < n; i++ {
		if inLine {
			if stmt[i] == '\n' {
				inLine = false
			}
			continue
		}
		if inBlock {
			if i+1 < n && stmt[i] == '*' && stmt[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if inSingle {
			builder.WriteByte(stmt[i])
			if stmt[i] == '\'' {
				if i+1 < n && stmt[i+1] == '\'' {
					builder.WriteByte(stmt[i+1])
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDouble {
			builder.WriteByte(stmt[i])
			if stmt[i] == '"' {
				inDouble = false
			}
			continue
		}
		if dollarTag != "" {
			builder.WriteByte(stmt[i])
			if strings.HasPrefix(stmt[i:], dollarTag) {
				for j := 1; j < len(dollarTag); j++ {
					builder.WriteByte(stmt[i+j])
				}
				i += len(dollarTag) - 1
				dollarTag = ""
			}
			continue
		}

		if i+1 < n && stmt[i] == '-' && stmt[i+1] == '-' {
			inLine = true
			i++
			continue
		}
		if i+1 < n && stmt[i] == '/' && stmt[i+1] == '*' {
			inBlock = true
			i++
			continue
		}
		if stmt[i] == '\'' {
			inSingle = true
			builder.WriteByte(stmt[i])
			continue
		}
		if stmt[i] == '"' {
			inDouble = true
			builder.WriteByte(stmt[i])
			continue
		}
		if stmt[i] == '$' {
			if tag, ok := matchDollarTag(stmt, i); ok {
				dollarTag = tag
				builder.WriteString(tag)
				i += len(tag) - 1
				continue
			}
		}
		builder.WriteByte(stmt[i])
	}
	return strings.TrimSpace(builder.String()) != ""
}
