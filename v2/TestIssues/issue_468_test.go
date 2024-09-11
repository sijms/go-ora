package TestIssues

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestIssue468(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_468(N NUMBER)`)
	}
	dropTable := func(db *sql.DB) error {
		return execCmd(db, "DROP TABLE TTB_468 PURGE")
	}
	insert := func(db *sql.DB) error {
		data := make([]int, 100)
		for index := range data {
			data[index] = index + 1
		}
		_, err := db.Exec("INSERT INTO TTB_468(N) VALUES (:1)", data)
		return err
	}
	query := func(db *sql.DB) error {
		var (
			n1, n2 int32
			cursor sql.Rows
		)
		rows, err := db.Query("select n,cursor(select TTB_468.n from dual) from TTB_468 order by n")
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		index := 0
		for rows.Next() {
			err = rows.Scan(&n1, &cursor)
			if err != nil {
				return err
			}
			for cursor.Next() {
				err = cursor.Scan(&n2)
				if err != nil {
					return err
				}
			}
			err = cursor.Close()
			if err != nil {
				return err
			}
			if n1 != n2 {
				return fmt.Errorf("n1: %d, n2: %d is not equal", n1, n2)
			}
			index++
		}
		t.Logf("%d rows scanned", index)
		return nil
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
	err = createTable(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = insert(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = query(db)
	if err != nil {
		t.Error(err)
		return
	}
}
