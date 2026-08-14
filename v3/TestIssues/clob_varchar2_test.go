package TestIssues

import (
	"database/sql"
	"strings"
	"testing"
)

func TestClobToVarchar2Conversion(t *testing.T) {
	var createTable = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE GOORA_TEMP_CLOB_VARCHAR2(
	ID	NUMBER(10)	NOT NULL,
	MSG	VARCHAR2(4000),
	PRIMARY KEY(ID)
)`)
	}

	var dropTable = func(db *sql.DB) error {
		return execCmd(db, "drop table GOORA_TEMP_CLOB_VARCHAR2 purge")
	}

	longText := strings.Repeat("status: 400|400 Bad Request|", 160)
	if len(longText) < 2000 {
		t.Fatalf("test text too short: %d bytes", len(longText))
	}

	var insert = func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO GOORA_TEMP_CLOB_VARCHAR2(ID, MSG) VALUES(1, :MSG)", longText)
		return err
	}

	var query = func(db *sql.DB) error {
		var (
			id  int
			msg string
		)
		err := db.QueryRow("SELECT ID, MSG FROM GOORA_TEMP_CLOB_VARCHAR2 WHERE ID = 1").Scan(&id, &msg)
		if err != nil {
			return err
		}
		if msg != longText {
			return &clobMismatchError{expected: longText, actual: msg}
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
}

func TestClobToVarchar2MultiByte(t *testing.T) {
	var createTable = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE GOORA_TEMP_CLOB_VARCHAR2_MB(
	ID	NUMBER(10)	NOT NULL,
	MSG	VARCHAR2(4000),
	PRIMARY KEY(ID)
)`)
	}

	var dropTable = func(db *sql.DB) error {
		return execCmd(db, "drop table GOORA_TEMP_CLOB_VARCHAR2_MB purge")
	}

	segment := "این یک متن تستی است برای بررسی رمزگذاری. "
	longText := strings.Repeat(segment, 80)
	if len([]rune(longText)) < 2000 {
		t.Fatalf("test text too short: %d runes", len([]rune(longText)))
	}

	var insert = func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO GOORA_TEMP_CLOB_VARCHAR2_MB(ID, MSG) VALUES(1, :MSG)", longText)
		return err
	}

	var query = func(db *sql.DB) error {
		var (
			id  int
			msg string
		)
		err := db.QueryRow("SELECT ID, MSG FROM GOORA_TEMP_CLOB_VARCHAR2_MB WHERE ID = 1").Scan(&id, &msg)
		if err != nil {
			return err
		}
		if msg != longText {
			return &clobMismatchError{expected: longText, actual: msg}
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
}
