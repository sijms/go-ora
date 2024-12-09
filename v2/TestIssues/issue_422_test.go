package TestIssues

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

func TestIssue422(t *testing.T) {
	type TimeDebug struct {
		CurrentDate          string    `db:"CURRENT_DATE"`
		CurrentTimestamp     string    `db:"CURRENT_TIMESTAMP"`
		Sysdate              string    `db:"SYSDATE"`
		Systimestamp         string    `db:"SYSTIMESTAMP"`
		TimeCurrentDate      time.Time `db:"T_CURRENT_DATE"`
		TimeCurrentTimestamp time.Time `db:"T_CURRENT_TIMESTAMP"`
		TimeSysdate          time.Time `db:"T_SYSDATE"`
		TimeSystimestamp     time.Time `db:"T_SYSTIMESTAMP"`
	}
	var dumpJson = func(d any) {
		b, _ := json.MarshalIndent(d, "", "  ")
		t.Logf("%s\n", string(b))
	}

	var dumpTime = func(r TimeDebug) {
		t.Logf("%16s: %s\n", "Sysdate", r.TimeSysdate.Format(time.UnixDate))
		t.Logf("%16s: %s\n", "Systimestamp", r.TimeSystimestamp.Format(time.UnixDate))
		t.Logf("%16s: %s\n", "Real Time", time.Now().Format(time.UnixDate))
	}

	var query = func(db *sql.DB) error {
		sqlText := `
        SELECT
            to_char(CURRENT_DATE, 'YYYY-MM-DD HH24:MI:SS') "CURRENT_DATE",
            to_char(CURRENT_TIMESTAMP, 'YYYY-MM-DD HH24:MI:SS TZR') "CURRENT_TIMESTAMP",
            to_char(SYSDATE, 'YYYY-MM-DD HH24:MI:SS') "SYSDATE",
            to_char(SYSTIMESTAMP, 'YYYY-MM-DD HH24:MI:SS TZR') "SYSTIMESTAMP",
            CURRENT_DATE as T_CURRENT_DATE,
            CURRENT_TIMESTAMP as T_CURRENT_TIMESTAMP,
            SYSDATE as T_SYSDATE,
            SYSTIMESTAMP as T_SYSTIMESTAMP
         FROM DUAL`
		var r = TimeDebug{}
		err := queryStruct(db.QueryRow(sqlText), &r)
		if err != nil {
			return err
		}
		dumpJson(r)
		dumpTime(r)
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
	err = query(db)
	if err != nil {
		t.Error(err)
		return
	}
}
