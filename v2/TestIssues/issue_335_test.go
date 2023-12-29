package TestIssues

import (
	"database/sql"
	"testing"
)

func TestIssue335(t *testing.T) {
	var create = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_335 (
	  varchar2_col VARCHAR2(10),
	  char_col CHAR(10),
	  nchar_col NCHAR(10),
	  number_col NUMBER(10, 2),
	  integer_col INTEGER,
	  smallint_col SMALLINT,
	  float_col FLOAT(10),
	  date_col DATE,
	  timestamp_col TIMESTAMP,
	  timestamp_with_time_zone_col TIMESTAMP WITH TIME ZONE,
	  timestamp_with_local_time_zone_col TIMESTAMP WITH LOCAL TIME ZONE,
	  interval_year_to_month_col INTERVAL YEAR TO MONTH,
	  interval_day_to_second_col INTERVAL DAY TO SECOND,
	  clob_col CLOB,
	  blob_col BLOB,
	  bfile_col BFILE,
	  long_raw_col LONG RAW,
	  raw_col RAW(10)
	)NOCOMPRESS`,
			`INSERT INTO TTB_335 (
		varchar2_col, 
		char_col, 
		nchar_col, 
		number_col, 
		integer_col, 
		smallint_col, 
		float_col, 
		date_col, 
		timestamp_col, 
		timestamp_with_time_zone_col, 
		timestamp_with_local_time_zone_col, 
		interval_year_to_month_col, 
		interval_day_to_second_col, 
		clob_col,
		blob_col,
		bfile_col,
		long_raw_col,
		raw_col)
		VALUES (
		'hello', 
		'world', 
		'unicode',
		3.14, 
		42, 10,
		3.14159, 
		SYSDATE,
		SYSTIMESTAMP,
		SYSTIMESTAMP, 
		SYSTIMESTAMP, 
		INTERVAL '1-2' YEAR TO MONTH, INTERVAL '1 12:30:00' DAY TO SECOND, 
		'lorem ipsum',
		EMPTY_BLOB(), 
		BFILENAME('MY_DIR', 'MY_FILE'), 
		EMPTY_BLOB(), 
		'DEADBEEF')`,
		)
	}

	var drop = func(db *sql.DB) error {
		return execCmd(db, `DROP TABLE TTB_335 PURGE`)
	}
	var query = func(db *sql.DB) error {
		rows, err := db.Query("SELECT * FROM TTB_335")
		if err != nil {
			return err
		}
		defer func() {
			err = rows.Close()
			if err != nil {
				t.Error(err)
			}
		}()
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
	err = create(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = drop(db)
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
