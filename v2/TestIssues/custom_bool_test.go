package TestIssues

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	go_ora "github.com/sijms/go-ora/v2"
)

type OracleBool bool

func (b OracleBool) Value() (driver.Value, error) {
	if b {
		return "J", nil
	}
	return "N", nil
}

func (b *OracleBool) Scan(value interface{}) error {
	if val, ok := value.(string); ok {
		*b = val == "J"
	} else {
		return errors.New("non string result")
	}
	return nil
}

func TestCustomBool(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, "CREATE TABLE TB_420(COL1  VARCHAR2(1))")
	}

	dropTable := func(db *sql.DB) error {
		return execCmd(db, "drop table TB_420 purge")
	}
	insert := func(db *sql.DB) error {
		var b OracleBool = true
		_, err := db.Exec("INSERT INTO TB_420(col1) VALUES(:myBool)", b)
		return err
	}

	query := func(db *sql.DB) error {
		var result string
		var result2 OracleBool
		err := db.QueryRow("SELECT col1, col1 FROM TB_420").Scan(&result, &result2)
		if err != nil {
			return err
		}
		if result != "J" {
			return fmt.Errorf("expected %s and got %s", "J", result)
		}
		if !result2 {
			return fmt.Errorf("expected true and got false")
		}
		return nil
	}

	query2 := func(db *sql.DB) error {
		var result string
		var result2 OracleBool
		_, err := db.Exec("BEGIN SELECT col1, col1 into :1, :2 FROM TB_420; END;", go_ora.Out{Dest: &result, Size: 10},
			go_ora.Out{Dest: &result2, Size: 10})
		if err != nil {
			return err
		}
		if result != "J" {
			return fmt.Errorf("expected %s and got %s", "J", result)
		}
		if !result2 {
			return fmt.Errorf("expected true and got false")
		}
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

	err = query2(db)
	if err != nil {
		t.Error(err)
		return
	}
}
