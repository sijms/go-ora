package TestIssues

import (
	"database/sql"
	"strings"
	"testing"
)

func TestClobCharsetRoundTrip(t *testing.T) {
	var createTable = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE GOORA_TEMP_CLOB_CHARSET(
	ID	number(10)	NOT NULL,
	DATA CLOB,
	PRIMARY KEY(ID)
	)`)
	}

	var dropTable = func(db *sql.DB) error {
		return execCmd(db, "drop table GOORA_TEMP_CLOB_CHARSET purge")
	}

	// Build a string > 2000 chars with Persian/Arabic multi-byte UTF-8 characters
	// mixed with ASCII to exercise the CLOB write/read encoding path.
	persianSegment := "این یک متن تستی است برای بررسی صحت رمزگذاری و رمزگشایی CLOB. "
	// Repeat to exceed 2000 characters
	longText := strings.Repeat(persianSegment, 40)
	if len([]rune(longText)) < 2000 {
		t.Fatalf("test text too short: %d runes", len([]rune(longText)))
	}

	var insert = func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO GOORA_TEMP_CLOB_CHARSET(ID, DATA) VALUES(1, :DATA)", longText)
		return err
	}

	var query = func(db *sql.DB) error {
		var (
			id   int
			data string
		)
		err := db.QueryRow("SELECT ID, DATA FROM GOORA_TEMP_CLOB_CHARSET WHERE ID = 1").Scan(&id, &data)
		if err != nil {
			return err
		}
		if data != longText {
			return &clobMismatchError{expected: longText, actual: data}
		}
		return nil
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

	err = createTable(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			t.Error(err)
		}
	}()

	err = insert(db)
	if err != nil {
		t.Error(err)
		return
	}

	err = query(db)
	if err != nil {
		t.Error(err)
		return
	}
}

type clobMismatchError struct {
	expected string
	actual   string
}

func (e *clobMismatchError) Error() string {
	expRunes := []rune(e.expected)
	actRunes := []rune(e.actual)
	maxLen := len(expRunes)
	if len(actRunes) > maxLen {
		maxLen = len(actRunes)
	}
	firstDiff := -1
	for i := 0; i < maxLen; i++ {
		var expR, actR rune
		if i < len(expRunes) {
			expR = expRunes[i]
		}
		if i < len(actRunes) {
			actR = actRunes[i]
		}
		if expR != actR {
			firstDiff = i
			break
		}
	}
	if firstDiff == -1 {
		return "clob data mismatch: lengths differ but no rune difference found"
	}
	return "clob data mismatch: first differing rune at index " + itoa(firstDiff)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := false
	if n < 0 {
		negative = true
		n = -n
	}
	var digits [20]byte
	pos := len(digits)
	for n > 0 {
		pos--
		digits[pos] = byte('0' + n%10)
		n /= 10
	}
	if negative {
		pos--
		digits[pos] = '-'
	}
	return string(digits[pos:])
}
