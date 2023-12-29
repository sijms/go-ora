package TestIssues

import "testing"

func TestIssue186(t *testing.T) {
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
	rows, err := db.Query(`SELECT ROWID FROM "DVSYS"."COMMAND_RULE$" OFFSET 0 ROWS FETCH NEXT 1000 ROWS ONLY`)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	var rowID string
	for rows.Next() {
		err = rows.Scan(&rowID)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(rowID)
	}
}
