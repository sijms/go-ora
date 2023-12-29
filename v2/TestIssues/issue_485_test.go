package TestIssues

import "testing"

// this issue occur when passsing input parameter as nested pointers
func TestIssue_485(t *testing.T) {
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
	s := "ra"
	ps := &s
	rows, err := db.Query("SELECT 1 FROM DUAL WHERE 'A' = :1", &ps)
	if err != nil {
		t.Error(err)
		return
	}
	err = rows.Close()
	if err != nil {
		t.Error()
		return
	}
}
