package go_ora_test

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"testing"

	_ "github.com/sijms/go-ora"
	"github.com/sijms/go-ora/converters"
)

var conn *sql.DB

func TestMain(m *testing.M) {
	var err error
	connStr := os.Getenv("GOORA_TESTDB")
	if connStr == "" {
		log.Fatal(fmt.Errorf("Provide  oracle server url in environment variable GOORA_TESTDB"))
	}
	conn, err = sql.Open("oracle", connStr)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	rc := m.Run()

	conn.Close()
	os.Exit(rc)
}
func TestSelectBindFloat(t *testing.T) {
	for _, tt := range converters.TestFloatValue {
		t.Run(tt.SelectText, func(t *testing.T) {
			query := fmt.Sprintf("select N, N||'' S from (select :1 N from dual)")
			stmt, err := conn.Prepare(query)
			if err != nil {
				t.Errorf("Query can't be prepared: %s", err)
				return
			}
			defer stmt.Close()

			rows, err := stmt.Query(tt.Float)
			if err != nil {
				t.Errorf("Query can't be run: %s", err)
				return
			}

			defer rows.Close()
			if !rows.Next() {
				t.Errorf("Query returns no record")
				return
			}

			var (
				got float64
				s   string
			)
			err = rows.Scan(&got, &s)
			if err != nil {
				t.Errorf("Query can't scan row: %s", err)
				return
			}

			var e float64
			if tt.Float != 0.0 {
				e = math.Abs((got - tt.Float) / tt.Float)
			} else {
				e = math.Abs(got - tt.Float)
			}

			if e > 1e-15 {
				t.Errorf("Query expecting: %v, got %v", tt.Float, got)
			}
		})
	}

}

func TestSelectBindInt(t *testing.T) {
	for _, tt := range converters.TestFloatValue {
		if tt.IsInteger {
			t.Run(tt.SelectText, func(t *testing.T) {
				query := fmt.Sprintf("select N, N||'' S from (select :1 N from dual)")
				stmt, err := conn.Prepare(query)
				if err != nil {
					t.Errorf("Query can't be prepared: %s", err)
					return
				}
				defer stmt.Close()

				rows, err := stmt.Query(tt.Integer)
				if err != nil {
					t.Errorf("Query can't be run: %s", err)
					return
				}

				defer rows.Close()
				if !rows.Next() {
					t.Errorf("Query returns no record")
					return
				}

				var (
					got int64
					s   string
				)
				err = rows.Scan(&got, &s)
				if err != nil {
					t.Errorf("Query can't scan row: %s", err)
					return
				}

				if got != tt.Integer {
					t.Errorf("Query expecting: %v, got %v", tt.Integer, got)
				}
			})
		}
	}
}
