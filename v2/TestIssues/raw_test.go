package TestIssues

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestRaw(t *testing.T) {
	insert := func(db *sql.DB) error {
		sqlText := `INSERT INTO TTB_MAIN(ID, NAME, VAL, LDATE, DATA) VALUES(:ID, :NAME, :VAL, :LDATE, :DATA)`
		length := 500
		type TempStruct struct {
			Id   int       `db:"ID"`
			Name string    `db:"NAME"`
			Val  float32   `db:"VAL"`
			Date time.Time `db:"LDATE"`
			Data []byte    `db:"DATA"`
		}
		args := make([]TempStruct, length)
		for x := 0; x < length; x++ {
			args[x] = TempStruct{
				Id:   x + 1,
				Name: strings.Repeat("*", 20),
				Val:  float32(length) / float32(x+1),
				Date: time.Now(),
				Data: bytes.Repeat([]byte{55}, 20),
			}
			if x > 0 && x%10 == 0 {
				args[x].Data = []byte{}
			}
		}
		_, err := db.Exec(sqlText, args)
		return err
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
	var got []byte
	err = db.QueryRow(`SELECT DATA FROM TTB_MAIN WHERE ID = 10`).Scan(&got)
	if err != nil {
		t.Error(err)
		return
	}
	expected := bytes.Repeat([]byte{55}, 20)
	if bytes.Compare(got, expected) != 0 {
		t.Errorf("expected %#v and got %#v", expected, got)
	}
}
