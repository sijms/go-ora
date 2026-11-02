package TestIssues

import (
	"testing"
)

// Issue 543: reading the LONG column COLUMN_EXPRESSION of the index expression views
// failed with ORA-03146. The test builds its own function-based index so it does not
// depend on a schema (such as APEX) being installed in the test database.
func TestIssue543(t *testing.T) {
	expected := "CASE  WHEN \"SESSION_ID\" IS NULL THEN \"SECURITY_GROUP_ID\" END "
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
	err = execCmd(db,
		"CREATE TABLE TTB_ISSUE_543(SESSION_ID NUMBER, SECURITY_GROUP_ID NUMBER)",
		"CREATE INDEX IDX_ISSUE_543 ON TTB_ISSUE_543(CASE WHEN SESSION_ID IS NULL THEN SECURITY_GROUP_ID END)")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = execCmd(db, "DROP TABLE TTB_ISSUE_543 PURGE")
		if err != nil {
			t.Error(err)
		}
	}()
	var result string
	err = db.QueryRow("select column_expression from user_ind_expressions where index_name = 'IDX_ISSUE_543' and column_position = 1").Scan(&result)
	if err != nil {
		t.Error(err)
		return
	}
	if result != expected {
		t.Errorf("expected: %s and got: %s", expected, result)
	}
}
