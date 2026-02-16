package TestIssues

import "testing"

func TestIssue486(t *testing.T) {
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
	var value string
	err = db.QueryRow("SELECT 'AAA안녕하세요' FROM dual").Scan(&value)
	if err != nil {
		t.Error(err)
	}
	if value != "AAA안녕하세요" {
		t.Errorf("expected %s and got %s", "AAA안녕하세요", value)
	}
}
