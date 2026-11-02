package go_ora

import "testing"

// The expected names are what v2's TNSType.String() returns for the same codes,
// so DatabaseTypeName() keeps its value when an application moves from v2 to v3.
func TestOracleTypeName(t *testing.T) {
	want := map[uint16]string{
		1: "NCHAR", 2: "NUMBER", 3: "SB1", 4: "FLOAT", 5: "NullStr", 6: "VarNum", 7: "TNSType(7)",
		8: "LONG", 9: "VARCHAR", 11: "ROWID", 12: "DATE", 15: "VarRaw", 21: "BFloat", 22: "BDouble",
		23: "RAW", 24: "LongRaw", 60: "TNSType(60)", 68: "UINT", 94: "LongVarChar", 95: "LongVarRaw",
		96: "CHAR", 97: "CHARZ", 100: "IBFloat", 101: "IBDouble", 102: "REFCURSOR", 108: "OCIXMLType",
		109: "XMLType", 110: "OCIRef", 112: "OCIClobLocator", 113: "OCIBlobLocator", 114: "OCIFileLocator",
		116: "ResultSet", 119: "TNSType(119)", 127: "TNSType(127)", 155: "OCIString", 156: "OCIDate",
		180: "TimeStampDTY", 181: "TimeStampTZ_DTY", 182: "IntervalYM_DTY", 183: "IntervalDS_DTY",
		186: "TimeTZ", 187: "TIMESTAMP", 188: "TimeStampTZ", 189: "IntervalYM", 190: "IntervalDS",
		208: "UROWID", 231: "TimeStampLTZ_DTY", 232: "TimeStampeLTZ", 252: "TNSType(252)", 999: "TNSType(999)",
	}
	for code, name := range want {
		if got := oracleTypeName(code); got != name {
			t.Errorf("oracleTypeName(%d) = %q, want %q", code, got, name)
		}
	}
}
