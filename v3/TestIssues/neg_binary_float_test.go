// issue 580
package TestIssues

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
)

func TestNegBinaryFloat(t *testing.T) {
	var create = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_580(
	ID			number(10)	NOT NULL,
	COL1		BINARY_FLOAT,
	COL2		BINARY_DOUBLE
	)`,
			`INSERT INTO TTB_580 (ID, COL1, COL2) VALUES (1, 1.1, 1.1)`,
			`INSERT INTO TTB_580 (ID, COL1, COL2) VALUES (2, -1.1, -1.1)`)
	}
	var drop = func(db *sql.DB) error {
		return execCmd(db, `DROP TABLE TTB_580 PURGE`)
	}
	var query = func(db *sql.DB) error {
		var (
			id         int
			col1, col2 float64
		)
		rows, err := db.Query("SELECT ID, COL1, COL2 FROM TTB_580")
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		for rows.Next() {
			if err := rows.Scan(&id, &col1, &col2); err != nil {
				return err
			}
			switch id {
			case 1:
				if col1 != 1.1 && col2 != 1.1 {
					return fmt.Errorf("expected 1.1, 1.1 got %v, %v", col1, col2)
				}
			case 2:
				if col1 != -1.1 && col2 != -1.1 {
					return fmt.Errorf("expected -1.1, -1.1 got %v, %v", col1, col2)
				}
			default:
				return errors.New("invalid id")
			}
		}
		return rows.Err()
	}
	db, err := getDB()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	err = create(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = drop(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = query(db)
	if err != nil {
		t.Error(err)
		return
	}
}
