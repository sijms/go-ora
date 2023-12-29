package TestIssues

import "testing"

func TestBinaryDouble(t *testing.T) {
	db, err := getDB()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			t.Error()
		}
	}()
	var result float64
	err = db.QueryRow(`SELECT cast(298.04 as binary_float) FROM dual`).Scan(&result)
	if err != nil {
		t.Error(err)
		return
	}
	if result != 298.04 {
		t.Errorf("expected %v and got %v", 298.04, result)
		return
	}
}
