package TestIssues

import (
	"database/sql"
	"testing"

	go_ora "github.com/sijms/go-ora/v2"
)

func TestEmoji(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TEST_EMOJI(
	ID	number(10)	NOT NULL,
	CONTENT		NCLOB
	)`)
	}
	dropTable := func(db *sql.DB) error {
		return execCmd(db, `DROP TABLE TEST_EMOJI PURGE`)
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

	expectedValue := "😁🍎$➔®≧①◎◉§❤️🇨🇳"
	var got string
	var id int
	_, err = db.Exec("INSERT INTO TEST_EMOJI(ID, CONTENT) VALUES(1, :1)", go_ora.NClob{String: expectedValue, Valid: true})
	if err != nil {
		t.Error(err)
		return
	}
	err = db.QueryRow("SELECT ID, CONTENT FROM TEST_EMOJI").Scan(&id, &got)
	if err != nil {
		t.Error(err)
		return
	}
	if expectedValue != got {
		t.Errorf("expected: %s and got %s", expectedValue, got)
	}
}
