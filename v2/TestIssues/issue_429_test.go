package TestIssues

import (
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestIssue429(t *testing.T) {
	type TTB_DATA struct {
		Id   int64     `db:"ID"`
		Name string    `db:"NAME"`
		Val  float64   `db:"VAL"`
		Date time.Time `db:"LDATE"`
	}
	var insert = func(db *sql.DB) error {
		data := make([]TTB_DATA, 100)
		for x, _ := range data {
			data[x].Id = int64(1000000000 + x)
			data[x].Name = "test_" + strconv.Itoa(x)
			data[x].Val = 100.23 + 1
			data[x].Date = time.Now()
		}
		_, err := db.Exec("INSERT INTO TTB_MAIN (ID, NAME, VAL, LDATE) VALUES(:ID, :NAME, :VAL, :LDATE)", data)
		if err != nil {
			return err
		}
		return nil
	}
	var query = func(db *sql.DB) error {
		var id int
		err := db.QueryRow("SELECT ID FROM TTB_MAIN WHERE NAME = :1", "test_0").Scan(&id)
		if err != nil {
			return err
		}
		if id != 1000000000 {
			return fmt.Errorf("expected: %d and got: %d", 1000000000, id)
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
