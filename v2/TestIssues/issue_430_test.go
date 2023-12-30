package TestIssues

import (
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestIssue430(t *testing.T) {
	var insert = func(db *sql.DB) error {
		type TTB_DATA struct {
			Id   int64     `db:"ID"`
			Name string    `db:"NAME"`
			Val  float64   `db:"VAL"`
			Date time.Time `db:"LDATE"`
		}
		data := make([]TTB_DATA, 100)
		for x, _ := range data {
			data[x].Id = int64(1 + x)
			data[x].Name = "test_" + strconv.Itoa(x)
			data[x].Val = 100.23 + 1
			data[x].Date = time.Now()
		}
		_, err := db.Exec("INSERT INTO TEMP_TABLE_357 (ID, NAME, VAL, LDATE) VALUES(:ID, :NAME, :VAL, :LDATE)", data)
		if err != nil {
			return err
		}
		return nil
	}
	var query = func(db *sql.DB) error {
		result := struct {
			Id   int64     `db:"ID,number,,output"`
			Name string    `db:"NAME,,200,output"`
			Val  float64   `db:"VAL,number,,output"`
			Date time.Time `db:"LDATE,,,output"`
		}{}
		_, err := db.Exec(`BEGIN
SELECT ID, NAME, VAL, LDATE INTO :ID, :NAME, :VAL, :LDATE FROM TEMP_TABLE_357 WHERE ID = 1;
END;`, &result)
		if err != nil {
			return err
		}
		if result.Val != 101.23 {
			return fmt.Errorf("expected: %f and got: %f", 100.23, result.Val)
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
