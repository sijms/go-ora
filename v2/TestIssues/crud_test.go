package TestIssues

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCrud(t *testing.T) {
	type INPUT struct {
		ID   int       `db:"ID"`
		Name string    `db:"NAME"`
		Val  float64   `db:"VAL"`
		Date time.Time `db:"LDATE"`
		Data []byte    `db:"DATA"`
	}

	rowCount := func(db *sql.DB) (count int, err error) {
		err = db.QueryRow("SELECT COUNT(*) FROM TTB_MAIN").Scan(&count)
		return
	}
	insert := func(db Execuer) error {
		// insert 10 rows
		temp := INPUT{}
		var err error
		for x := 0; x < 10; x++ {
			temp.ID = x + 1
			temp.Name = strings.Repeat("-", x)
			temp.Val += 1.1
			temp.Date = time.Now()
			temp.Data = bytes.Repeat([]byte{55}, x)
			_, err = db.Exec("INSERT INTO TTB_MAIN(ID, NAME, VAL, LDATE, DATA) VALUES(:ID, :NAME, :VAL, :LDATE, :DATA)", &temp)
			if err != nil {
				return err
			}
		}
		return nil
	}
	insertWithPrepare := func(db Execuer) error {
		temp := INPUT{}
		stmt, err := db.Prepare("INSERT INTO TTB_MAIN(ID, NAME, VAL, LDATE, DATA) VALUES(:ID, :NAME, :VAL, :LDATE, :DATA)")
		if err != nil {
			return err
		}
		defer func() {
			err = stmt.Close()
			if err != nil {
				t.Error(err)
			}
		}()
		for x := 0; x < 10; x++ {
			temp.ID = x + 1
			temp.Name = strings.Repeat("-", x+1)
			temp.Val += 1.1
			temp.Date = time.Now()
			temp.Data = bytes.Repeat([]byte{55}, x+1)
			_, err = stmt.Exec(&temp)
			if err != nil {
				return err
			}
		}
		return nil
	}
	insertBulk := func(db Execuer) error {
		data := make([]INPUT, 100)
		baseVal := 1.1
		for index := range data {
			data[index].ID = index + 1
			data[index].Name = strings.Repeat("-", index+1)
			data[index].Val = baseVal + float64(index)
			data[index].Date = time.Now()
			data[index].Data = bytes.Repeat([]byte{55}, index+1)
		}
		_, err := db.Exec("INSERT INTO TTB_MAIN(ID, NAME, VAL, LDATE, DATA) VALUES(:ID, :NAME, :VAL, :LDATE, :DATA)",
			data)
		return err
	}
	transaction := func(db *sql.DB, dbFunc func(db Execuer) error, commit bool, rowsIncrement int) error {
		initialCount, err := rowCount(db)
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		err = dbFunc(tx)
		if err != nil {
			return err
		}
		if commit {
			err = tx.Commit()
		} else {
			err = tx.Rollback()
		}
		if err != nil {
			return err
		}
		count, err := rowCount(db)
		if err != nil {
			return err
		}
		if commit {
			if count != initialCount+rowsIncrement {
				return fmt.Errorf("before insert: %d rows and after insert with commit: %d rows",
					initialCount, count)
			}
		} else {
			if count != initialCount {
				return fmt.Errorf("before insert: %d rows and after insert with rollback: %d rows",
					initialCount, count)
			}
		}
		return nil
	}
	delRows := func(db *sql.DB) error {
		return execCmd(db, "DELETE FROM TTB_MAIN")
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
	err = delRows(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = insertWithPrepare(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = delRows(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = insertBulk(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = delRows(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = transaction(db, insertWithPrepare, false, 0)
	if err != nil {
		t.Error(err)
		return
	}
	err = transaction(db, insertBulk, false, 0)
	if err != nil {
		t.Error(err)
		return
	}
	err = transaction(db, insert, true, 10)
	if err != nil {
		t.Error(err)
		return
	}
	err = delRows(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = transaction(db, insertBulk, true, 100)
	if err != nil {
		t.Error(err)
		return
	}
	err = delRows(db)
	if err != nil {
		t.Error(err)
		return
	}
}
