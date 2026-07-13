package firebird

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/nakagami/firebirdsql"

	"ticker/internal/config"
)

func DSN(cfg config.Config) string {
	// firebirdsql DSN examples commonly use:
	//   user:pass@host:port/dbpath
	// Database path is the remote Firebird database path.
	return fmt.Sprintf("%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
}

var identRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

func normalizeIdent(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("identifier is empty")
	}
	up := strings.ToUpper(s)
	if !identRe.MatchString(up) {
		return "", fmt.Errorf("identifier %q must match %s (letters/digits/underscore, starting with a letter)", s, identRe.String())
	}
	return up, nil
}

func EnsureTable(ctx context.Context, db *sql.DB, table, column string) (string, string, error) {
	tbl, err := normalizeIdent(table)
	if err != nil {
		return "", "", fmt.Errorf("table: %w", err)
	}
	col, err := normalizeIdent(column)
	if err != nil {
		return "", "", fmt.Errorf("column: %w", err)
	}

	// Firebird stores identifiers uppercased and padded in RDB$RELATION_NAME.
	var one int
	q := `SELECT 1 FROM RDB$RELATIONS WHERE TRIM(RDB$RELATION_NAME) = ? AND (RDB$SYSTEM_FLAG IS NULL OR RDB$SYSTEM_FLAG = 0)`
	err = db.QueryRowContext(ctx, q, tbl).Scan(&one)
	if err == nil {
		return tbl, col, nil
	}
	if err != sql.ErrNoRows {
		return "", "", fmt.Errorf("check existing table: %w", err)
	}

	ddl := fmt.Sprintf("CREATE TABLE %s (%s TIMESTAMP)", tbl, col)
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return "", "", fmt.Errorf("create table: %w", err)
	}
	return tbl, col, nil
}

func InsertCurrentTimestampOnce(ctx context.Context, cfg config.Config) error {
	db, err := sql.Open("firebirdsql", DSN(cfg))
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	tbl, col, err := EnsureTable(ctx, db, cfg.Table, cfg.Column)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (CURRENT_TIMESTAMP)", tbl, col)
	if _, err := tx.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

