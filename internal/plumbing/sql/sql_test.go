package sql

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/olexsmir/viye/internal/viye"
	"olexsmir.xyz/x/is"
)

func TestMatch(t *testing.T) {
	for tname, tt := range map[string]struct {
		cmd  string
		args []string
		want bool
	}{
		"select":                    {"select", []string{"*", "from", "users"}, true},
		"select with no expression": {"select", nil, false},
		"insert":                    {"insert", []string{"into", "users"}, true},
		"insert without into":       {"insert", []string{"users"}, false},
		"update":                    {"update", []string{"users", "set", "x=1"}, true},
		"update without set":        {"update", []string{"users", "x=1"}, false},
		"delete":                    {"delete", []string{"from", "users"}, true},
		"delete missing table":      {"delete", []string{"from"}, false},
		"create":                    {"create", []string{"table", "t"}, true},
		"create incomplete":         {"create", []string{"table"}, false},
		"drop":                      {"drop", []string{"table", "t"}, true},
		"alter":                     {"alter", []string{"table", "t"}, true},
		"pragma":                    {"pragma", []string{"table_info(t)"}, true},
		"explain":                   {"explain", []string{"select", "1"}, true},
		"analyze alone":             {"analyze", nil, true},
		"uppercase":                 {"SELECT", []string{"*", "from", "t"}, true},
		"not sql":                   {"curl", []string{"https://x"}, false},
	} {
		t.Run(tname, func(t *testing.T) {
			if got := (Tool{}).Match(&viye.Context{Cmd: tt.cmd, Args: tt.args}); got != tt.want {
				t.Errorf("Match(%q, %v) = %v, want %v", tt.cmd, tt.args, got, tt.want)
			}
		})
	}
}

func TestRunQuery_select(t *testing.T) {
	db := newTestDB(t)
	_, err := db.Exec("insert into users values (1, 'alice', 30), (2, 'bob', NULL)")
	is.Err(t, err, nil)

	want := `: id  name   age
: 1   alice  30
: 2   bob    NULL
`
	for _, query := range []string{
		"select * from users order by id",
		"SELECT * FROM users ORDER BY id",
	} {
		out, err := runQuery(t.Context(), db, query)
		is.Err(t, err, nil)
		is.Equal(t, want, out)
	}
}

func TestRunQuery_truncation(t *testing.T) {
	db := newTestDB(t)
	_, err := db.ExecContext(t.Context(), "with recursive c(x) as (select 1 union all select x+1 from c where x < 105) insert into users select x, 'u'||x, x from c")
	is.Err(t, err, nil)

	out, err := runQuery(t.Context(), db, "select * from users")
	is.Err(t, err, nil)
	if !strings.Contains(out, "... 100+ more row(s), output truncated at 100 rows") {
		t.Errorf("output not truncated:\n%s", out)
	}
}

func TestRunQuery_write(t *testing.T) {
	db := newTestDB(t)
	_, err := db.ExecContext(t.Context(), "insert into users values (1, 'alice', 30)")
	is.Err(t, err, nil)

	for _, tt := range []struct{ query, want string }{
		{"update users set age = 40 where id = 1", "| 1 row(s) affected"},
		{"insert into users values (3, 'carol', 40)", "| 1 row(s) affected"},
		{"delete from users where id = 3", "| 1 row(s) affected"},
	} {
		out, err := runQuery(t.Context(), db, tt.query)
		is.Err(t, err, nil)
		is.Equal(t, tt.want, out)
	}
}

func TestUpdateFromBody(t *testing.T) {
	db := newTestDB(t)
	_, err := db.ExecContext(t.Context(), "insert into users values (1, 'alice', 30)")
	is.Err(t, err, nil)

	out, err := updateFromBody(t.Context(), db, "sqlite", "update users",
		[]string{"id name age", "1 alice 31", "3 carol 40", "garbage"})
	is.Err(t, err, nil)
	is.Equal(t, "| 1 updated, 1 inserted, 1 skipped (bad column count)", out)

	var count, aliceAge int
	err = db.QueryRowContext(t.Context(), "select count(*) from users").Scan(&count)
	is.Err(t, err, nil)
	is.Equal(t, 2, count)

	err = db.QueryRowContext(t.Context(), "select age from users where id = 1").Scan(&aliceAge)
	is.Err(t, err, nil)
	is.Equal(t, 31, aliceAge)
}

func TestExtractTableName(t *testing.T) {
	for _, tt := range []struct {
		name, query, want string
		ok                bool
	}{
		{"update", "update users set name = 'x' where id = 1", "users", true},
		{"update with subquery", "update users set n = (select 1 from roles) where id = 1", "users", true},
		{"select", "select * from users;", "users", true},
		{"insert", "insert into logs (a) values (1)", "logs", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := extractTableName(tt.query)
			is.Equal(t, tt.want, got)
			is.Equal(t, tt.ok, ok)
		})
	}
}

func TestDriverForDSN(t *testing.T) {
	for _, tt := range []struct {
		dsn, want string
		err       any
	}{
		{"postgres://user:pass@host/db", "postgres", nil},
		{"host=db user=alice", "postgres", nil},
		{"postgres://host/name.db", "postgres", nil},
		{"/tmp/data.db", "sqlite", nil},
		{"/tmp/app.sqlite3", "sqlite", nil},
		{"mongodb://host/db", "", "cannot determine driver"},
	} {
		got, err := driverForDSN(tt.dsn)
		is.Err(t, err, tt.err)
		is.Equal(t, tt.want, got)
	}
}

func TestPlaceholder(t *testing.T) {
	is.Equal(t, "$3", placeholder("postgres", 2))
	for _, i := range []int{0, 3} {
		is.Equal(t, "?", placeholder("sqlite", i))
	}
}

func TestWithLimit(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"select * from users;", "select * from users limit 101"},
		{"select * from users limit 10", "select * from users limit 10"},
		{"select * from users limit 10 offset 5", "select * from users limit 10 offset 5"},
		{"pragma table_info(users)", "pragma table_info(users)"},
		{"update users set age = 1", "update users set age = 1"},
	} {
		is.Equal(t, tt.want, withLimit(tt.in))
	}
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	is.Err(t, err, nil)
	db.SetMaxOpenConns(1) // keep connection shared across queries.
	t.Cleanup(func() { db.Close() })
	_, err = db.ExecContext(t.Context(), "create table users (id integer primary key, name text, age integer)")
	is.Err(t, err, nil)
	return db
}
