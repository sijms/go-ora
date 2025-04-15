package TestIssues

import (
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"
)

type ValuerExample struct {
	TheDate int64
}

func (ve *ValuerExample) Value() (driver.Value, error) {
	return time.Unix(ve.TheDate, 0), nil
}

func TestIssue662(t *testing.T) {
	ve := &ValuerExample{
		TheDate: 1744732127,
	}
	var create = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_662(
    THEDATE          date not null
	)`)
	}
	var drop = func(db *sql.DB) error {
		return execCmd(db,
			"DROP TABLE TTB_662",
		)
	}

	var insert = func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO TTB_662(THEDATE) VALUES(:1)", ve)
		if err != nil {
			return err
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
	err = insert(db)
	if err != nil {
		t.Error(err)
		return
	}
}
