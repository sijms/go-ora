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
			query := fmt.Sprintf("select :1 N from dual")
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
			)
			err = rows.Scan(&got)
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
				t.Errorf("Query expecting: %v, got %v, error %g", tt.Float, got, e)
			}
		})
	}

}

func TestSelectBindInt(t *testing.T) {
	for _, tt := range converters.TestFloatValue {
		if tt.IsInteger {
			t.Run(tt.SelectText, func(t *testing.T) {
				query := fmt.Sprintf("select :1 N from dual")
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
				)
				err = rows.Scan(&got)
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

func TestSelectBindFloatAsInt(t *testing.T) {
	for _, tt := range converters.TestFloatValue {
		t.Run(tt.SelectText, func(t *testing.T) {
			query := fmt.Sprintf("select :1 N from dual")
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
				got int64
			)
			err = rows.Scan(&got)

			if err == nil && !tt.IsInteger {
				t.Errorf("Expecting an error when scanning a real float(%v) into an int", tt.Float)
				return
			}

			if err != nil && tt.IsInteger {
				t.Errorf("Un-expecting an error when scanning a int as a float(%v) into an int, err:%s", tt.Float, err)
				return
			}

			if err != nil {
				return
			}

			if got != tt.Integer {
				t.Errorf("Expecting: int64(%v), got int64(%v),%v", int64(float64(tt.Integer)), got, tt.Float)
			}
		})
	}
}

func TestSelectBindIntAsFloat(t *testing.T) {
	for _, tt := range converters.TestFloatValue {
		if tt.IsInteger {
			t.Run(tt.SelectText, func(t *testing.T) {
				query := fmt.Sprintf("select :1 N from dual")
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
					got float64
				)
				err = rows.Scan(&got)

				if err == nil && !tt.IsInteger {
					t.Errorf("Expecting an error when scanning a real float(%v) into an int", tt.Float)
					return
				}

				if tt.Float != 0.0 {
					e := math.Abs((got - tt.Float) / tt.Float)
					if e > 1e-15 {
						t.Errorf("DecodeDouble(EncodeDouble(%g)) = %g,  Diff= %e", tt.Float, got, e)
					}
				} else if got != 0.0 {
					t.Errorf("DecodeDouble(EncodeDouble(%g)) = %g", tt.Float, got)

				}
			})
		}
	}
}
