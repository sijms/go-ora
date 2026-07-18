package utils

import "database/sql"

type (
	Execuer interface {
		Exec(query string, args ...any) (sql.Result, error)
		Prepare(query string) (*sql.Stmt, error)
	}
)
