/* -*- coding: utf-8-unix -*- */

package tests

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora"
	"os"
	"testing"
	"time"
)

func TestQueryWithSqlContainingZhs(t *testing.T) {
	dsn := os.Getenv("GOORA_TESTDB")
	fmt.Println(dsn)
	db, err := sql.Open("oracle", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin: %v", err)
		return
	}
	defer tx.Rollback()

	cmdText := `select sysdate 中文 from dual`

	ctx := context.TODO()
	rows, err := tx.QueryContext(ctx, cmdText)
	if err != nil {
		t.Fatalf("tx.QueryContext: %v", err)
		return
	}

	fmt.Printf("%T", rows)

	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("rows.Columns: %v", err)
		return
	} else {
		if cols[0] != "中文" {
			t.Fatalf("expecting '中文' but got: '%s'.", cols[0])
		}
	}

	for rows.Next() {
		now := time.Unix(0, 0)
		if err := rows.Scan(&now); err != nil {
			t.Fatalf("rows.Scan: %v", err)
		} else {
			fmt.Printf("now: %v", now)
		}
	}
}
