package TestIssues

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func TestIssue287(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_287(
	TS TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	}
	dropTable := func(db *sql.DB) error {
		return execCmd(db, "DROP TABLE TTB_287 PURGE")
	}
	insert := func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO TTB_287 VALUES(DEFAULT)")
		return err
	}
	query := func(db *sql.DB) error {
		rows, err := db.Query("SELECT * FROM TTB_287")
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		values := make([]any, len(cols))
		valuesWrapped := make([]any, 0, len(cols))
		for i := range cols {
			valuesWrapped = append(valuesWrapped, &values[i])
		}
		_ = rows.Next()
		err = rows.Scan(valuesWrapped...)
		if err != nil {
			return err
		}
		jObj := map[string]any{}
		for i, v := range values {
			jObj[cols[i]] = v
		}
		b, _ := json.Marshal(jObj["TS"])
		t.Log(string(b))
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
