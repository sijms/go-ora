package go_ora

import "testing"

// 29-column point lookup, a typical read-path statement (~370 characters)
const newStmtBenchSQL = `SELECT COLUMN_01, COLUMN_02, COLUMN_03, COLUMN_04, COLUMN_05, COLUMN_06, COLUMN_07, COLUMN_08, COLUMN_09, COLUMN_10, COLUMN_11, COLUMN_12, COLUMN_13, COLUMN_14, COLUMN_15, COLUMN_16, COLUMN_17, COLUMN_18, COLUMN_19, COLUMN_20, COLUMN_21, COLUMN_22, COLUMN_23, COLUMN_24, COLUMN_25, COLUMN_26, COLUMN_27, COLUMN_28, COLUMN_29 FROM APP.ORDERS WHERE ORDER_ID = :orderId`

func BenchmarkNewStmtSelect29Cols(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewStmt(newStmtBenchSQL, nil)
	}
}

func TestNewStmtReturnClause(t *testing.T) {
	cases := []struct {
		sql  string
		want bool
	}{
		{newStmtBenchSQL, false},
		{"UPDATE T SET A = :a WHERE ID = :id RETURNING B INTO :b", true},
		{"update t set a = :a returning b, c into :b, :c", true},
		{"DELETE FROM T WHERE ID = :id RETURN A INTO :a", true},
		{"INSERT INTO T (A) SELECT A FROM S", false},
		{"INSERT INTO T (A) VALUES (:a)", false},
		{"UPDATE T SET A = 'RETURNING x INTO y' WHERE ID = 1", false},
	}
	for _, c := range cases {
		if got := NewStmt(c.sql, nil)._hasReturnClause; got != c.want {
			t.Errorf("%q: _hasReturnClause = %v, want %v", c.sql, got, c.want)
		}
	}
}
