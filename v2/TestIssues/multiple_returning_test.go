package TestIssues

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestMultipleReturning(t *testing.T) {
	createTable := func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_329(
    ID          NUMBER(10),
    NAME VARCHAR2(100),
    TEAM_NAME VARCHAR2(100),
    ONBOARD_DATE DATE
	)`)
	}
	dropTable := func(db *sql.DB) error {
		return execCmd(db, "DROP TABLE TTB_329 PURGE")
	}
	insert := func(db *sql.DB) error {
		temp := struct {
			Id   int       `db:"ID"`
			Name string    `db:"NAME"`
			Team string    `db:"TEAM"`
			Date time.Time `db:"LDATE"`
			Ret1 string    `db:"RET1,,100,output"`
			Ret2 string    `db:"RET2,,100,output"`
		}{1, "test", "team", time.Now(), "", ""}
		_, err := db.Exec(`INSERT INTO TTB_329(ID, NAME, TEAM_NAME, ONBOARD_DATE)
VALUES(:ID, :NAME, :TEAM, :LDATE) RETURNING NAME, TEAM_NAME INTO :RET1, :RET2`, &temp)
		if err != nil {
			return err
		}
		if temp.Ret1 != "test" {
			return fmt.Errorf("expected: name=%s and got name=%s", "test", temp.Ret1)
		}
		if temp.Ret2 != "team" {
			return fmt.Errorf("expected: team=%s and got team=%s", "team", temp.Ret2)
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
}
