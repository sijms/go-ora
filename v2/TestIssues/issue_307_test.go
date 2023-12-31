package TestIssues

import (
	go_ora "github.com/sijms/go-ora/v2"
	"testing"
	"time"
)

func TestIssue307(t *testing.T) {
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
	err = createMainTable(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropMainTable(db)
		if err != nil {
			t.Error(err)
		}
	}()
	var id int
	_, err = db.Exec("INSERT INTO TTB_MAIN(ID, NAME, LDATE) VALUES(:1, :2, :3) RETURNING ID INTO :4",
		1, "TEST", time.Now(), go_ora.Out{Dest: &id})
	if err != nil {
		t.Error(err)
		return
	}
	if id != 1 {
		t.Errorf("expected: %d and got: %d", 1, id)
	}
}
