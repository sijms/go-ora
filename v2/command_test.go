package go_ora

import "testing"

func TestNewStmt_WithComments(t *testing.T) {
	t.Run("SELECT", func(t *testing.T) {
		querySelectWithComments := `
-- comment #1
  -- comment #2
/* comment #3 */
  /* comment #4 */ select * from dual
`

		stmt := NewStmt(querySelectWithComments, nil)
		if stmt == nil {
			t.Errorf("no stmt returned")
		} else if stmt.stmtType != SELECT {
			t.Errorf("expected stmt.stmtType to be %v but was %v", SELECT, stmt.stmtType)
		}
	})

	t.Run("UPDATE", func(t *testing.T) {
		querySelectWithComments := `
-- comment #1
  -- comment #2
/* comment #3 */
  /* comment #4 */ update foo set bar = 1 where baz = 1
`

		stmt := NewStmt(querySelectWithComments, nil)
		if stmt == nil {
			t.Errorf("no stmt returned")
		} else if stmt.stmtType != DML {
			t.Errorf("expected stmt.stmtType to be %v but was %v", DML, stmt.stmtType)
		}
	})

	t.Run("DECLARE", func(t *testing.T) {
		querySelectWithComments := `
-- comment #1
  -- comment #2
/* comment #3 */
  /* comment #4 */ 
DECLARE
   foo NUMBER := 42;
BEGIN
   INSERT INTO bar VALUES (foo);
END;
`

		stmt := NewStmt(querySelectWithComments, nil)
		if stmt == nil {
			t.Errorf("no stmt returned")
		} else if stmt.stmtType != PLSQL {
			t.Errorf("expected stmt.stmtType to be %v but was %v", PLSQL, stmt.stmtType)
		}
	})
}
