package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/olexsmir/viye/internal/config"
	"github.com/olexsmir/viye/internal/viye"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

type Tool struct{}

func (Tool) Name() string { return "sql(select, insert, update, delete, etc...)" }
func (Tool) Match(c *viye.Context) bool {
	switch strings.ToLower(c.Cmd) {
	case "select", "explain", "pragma":
		return len(c.Args) >= 1
	case "insert":
		return len(c.Args) >= 2 && strings.EqualFold(c.Args[0], "into")
	case "update":
		return len(c.Args) >= 2 && strings.EqualFold(c.Args[1], "set")
	case "delete":
		return len(c.Args) >= 2 && strings.EqualFold(c.Args[0], "from")
	case "create", "drop", "alter":
		return len(c.Args) >= 2
	case "analyze":
		return true
	}
	return false
}

func (Tool) Execute(c *viye.Context) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), viye.Timeout)
	defer cancel()

	query := strings.Join(append([]string{c.Cmd}, c.Args...), " ")
	db, driver, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	if len(c.Body) > 0 {
		return updateFromBody(ctx, db, driver, query, c.Body)
	}
	return runQuery(ctx, db, query)
}

const maxRows = 100

func runQuery(ctx context.Context, db *sql.DB, query string) (string, error) {
	if !isSelect(query) {
		r, err := db.ExecContext(ctx, query)
		if err != nil {
			return "", fmt.Errorf("sql: %s", err)
		}
		n, _ := r.RowsAffected()
		return fmt.Sprintf("| %d row(s) affected", n), nil
	}

	rows, err := db.QueryContext(ctx, withLimit(query))
	if err != nil {
		return "", fmt.Errorf("sql: %s", err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("sql: %s", err)
	}

	colWidths := make([]int, len(cols))
	for i, c := range cols {
		colWidths[i] = len(c)
	}

	src := make([]sql.NullString, len(cols))
	ptrs := make([]any, len(cols))
	for i := range src {
		ptrs[i] = &src[i]
	}

	var rowsData [][]string
	for rows.Next() && len(rowsData) <= maxRows {
		if err := rows.Scan(ptrs...); err != nil {
			return "", fmt.Errorf("sql: scan: %w", err)
		}
		r := make([]string, len(cols))
		for i, v := range src {
			if v.Valid {
				r[i] = v.String
			} else {
				r[i] = "NULL"
			}
			if len(r[i]) > colWidths[i] {
				colWidths[i] = len(r[i])
			}
		}
		rowsData = append(rowsData, r)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("sql: rows: %w", err)
	}

	var buf strings.Builder
	writeRow := func(fields []string) {
		buf.WriteString(": ")
		for i, f := range fields {
			buf.WriteString(f)
			if i < len(fields)-1 {
				buf.WriteString(strings.Repeat(" ", colWidths[i]-len(f)+2))
			}
		}
		buf.WriteByte('\n')
	}

	truncated := len(rowsData) > maxRows
	if truncated {
		rowsData = rowsData[:maxRows]
	}

	writeRow(cols)
	for _, r := range rowsData {
		writeRow(r)
	}
	if truncated {
		fmt.Fprintf(&buf, "... %d+ more row(s), output truncated at %d rows\n", maxRows, maxRows)
	}

	return buf.String(), nil
}

func updateFromBody(ctx context.Context, db *sql.DB, driver, query string, body []string) (string, error) {
	if len(body) < 2 {
		return "", errors.New("sql: body must have header + at least one data row")
	}
	header := strings.Fields(body[0])
	if len(header) == 0 {
		return "", fmt.Errorf("sql: invalid header line")
	}
	pkCol := header[0]
	table, ok := extractTableName(query)
	if !ok {
		return "", fmt.Errorf("sql: could not determine table name")
	}

	var updated, inserted, skipped int
	for _, line := range body[1:] {
		vals := strings.Fields(line)
		if len(vals) != len(header) {
			skipped++
			continue
		}

		existsQuery := fmt.Sprintf("select exists(select 1 from %s where %s = %s)", table, pkCol, placeholder(driver, 0))
		var exists bool
		if err := db.QueryRowContext(ctx, existsQuery, vals[0]).Scan(&exists); err != nil {
			return "", fmt.Errorf("sql: check exists: %w", err)
		}
		if exists {
			var sets []string
			var args []any
			for i, col := range header {
				if col == pkCol {
					continue
				}
				sets = append(sets, fmt.Sprintf("%s = %s", col, placeholder(driver, len(args))))
				args = append(args, vals[i])
			}
			args = append(args, vals[0])
			updateQuery := fmt.Sprintf("update %s set %s where %s = %s", table, strings.Join(sets, ", "), pkCol, placeholder(driver, len(args)-1))
			if _, err := db.ExecContext(ctx, updateQuery, args...); err != nil {
				return "", fmt.Errorf("sql: update: %w", err)
			}
			updated++
		} else {
			phs := make([]string, len(header))
			args := make([]any, len(header))
			for i := range header {
				phs[i] = placeholder(driver, i)
				args[i] = vals[i]
			}
			insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(header, ", "), strings.Join(phs, ", "))
			if _, err := db.ExecContext(ctx, insertQuery, args...); err != nil {
				return "", fmt.Errorf("sql: insert: %w", err)
			}
			inserted++
		}
	}

	msg := fmt.Sprintf("| %d updated, %d inserted", updated, inserted)
	if skipped > 0 {
		msg += fmt.Sprintf(", %d skipped (bad column count)", skipped)
	}
	return msg, nil
}

func placeholder(driver string, i int) string {
	if driver == "postgres" {
		return fmt.Sprintf("$%d", i+1)
	}
	return "?"
}

func isSelect(query string) bool {
	q := strings.ToLower(query)
	return strings.HasPrefix(q, "select") || strings.HasPrefix(q, "explain") ||
		strings.HasPrefix(q, "pragma")
}

func withLimit(query string) string {
	q := strings.TrimRight(strings.TrimSpace(query), "; \t\n")
	if !isSelect(q) || strings.HasPrefix(strings.ToLower(q), "pragma") || hasLimit(q) {
		return query
	}
	return q + fmt.Sprintf(" limit %d", maxRows+1)
}

func hasLimit(q string) bool {
	f := strings.Fields(q)
	for i := len(f) - 1; i >= 0 && i >= len(f)-4; i-- {
		if strings.EqualFold(f[i], "limit") {
			return true
		}
	}
	return false
}

func extractTableName(query string) (string, bool) {
	q := strings.ToUpper(strings.TrimSpace(query))
	for _, kw := range []string{"UPDATE ", "INTO ", "FROM "} {
		idx := strings.Index(q, kw)
		if idx >= 0 {
			parts := strings.Fields(query[idx+len(kw):])
			if len(parts) > 0 {
				return strings.Trim(parts[0], "\"`'[];"), true
			}
		}
	}
	return "", false
}

func openDB() (*sql.DB, string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, "", err
	}
	dsn, ok := cfg.Get("dsn")
	if !ok {
		return nil, "", errors.New("dsn option is not set")
	}

	driver, err := driverForDSN(dsn)
	if err != nil {
		return nil, "", err
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, "", err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, "", err
	}
	return db, driver, nil
}

func driverForDSN(dsn string) (string, error) {
	switch {
	case strings.HasPrefix(dsn, "postgres://") || strings.Contains(dsn, "user="):
		return "postgres", nil
	case strings.Contains(dsn, ".db") || strings.Contains(dsn, ".sqlite"):
		return "sqlite", nil
	default:
		return "", fmt.Errorf("sql: cannot determine driver from dsn %q", dsn)
	}
}
