package TestIssues

import (
	"database/sql"
	"testing"
	"time"
)

func TestIssue343(t *testing.T) {
	var createTable = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_343 
   (	"ID" NUMBER(20,0) NOT NULL ENABLE, 
		"TM" VARCHAR2(30), 
		"SN" VARCHAR2(25) NOT NULL ENABLE, 
		"CUS" VARCHAR2(20) NOT NULL ENABLE, 
		"AID" VARCHAR2(20), 
		"TR" VARCHAR2(8) NOT NULL ENABLE, 
		"PID" VARCHAR2(20), 
		"CODE" VARCHAR2(20) NOT NULL ENABLE, 
		"TTNO" VARCHAR2(20), 
		"UPDATETIME" TIMESTAMP (6) DEFAULT systimestamp, 
		 PRIMARY KEY ("ID")
	)`)
	}
	var dropTable = func(db *sql.DB) error {
		return execCmd(db, "DROP TABLE TTB_343 PURGE")
	}
	var merge = func(db *sql.DB) error {
		sqlText := `MERGE INTO TTB_343 t1 USING(select :ID ID from dual) tmp ON (tmp.ID=t1.ID) 
    WHEN MATCHED THEN UPDATE SET TM=:TM,SN=:SN,CUS=:CUS, AID=:AID,TR=:TR,PID=:PID,CODE=:CODE,TTNO=:TTNO,UPDATETIME=:UPDATETIME WHERE t1.ID=:ID AND t1.UPDATETIME<=:UPDATETIME 
    WHEN NOT MATCHED THEN INSERT (ID,TM,SN,CUS,AID,TR,PID,CODE,TTNO,UPDATETIME) VALUES (:ID,:TM,:SN,:CUS,:AID,:TR,:PID,:CODE,:TTNO,:UPDATETIME)`
		length := 500
		type TestShort struct {
			Id         int            `db:"ID"`
			Tm         sql.NullString `db:"TM"`
			Sn         string         `db:"SN"`
			Cus        string         `db:"CUS"`
			Aid        sql.NullString `db:"AID"`
			Tr         string         `db:"TR"`
			Pid        string         `db:"PID"`
			Code       string         `db:"CODE"`
			Ttno       sql.NullString `db:"TTNO"`
			UpdateTime time.Time      `db:"UPDATETIME,timestamp"`
		}
		data := make([]TestShort, length)
		for index, _ := range data {
			data[index].Id = index + 1
			if index > 0 && index%10 == 0 {
				data[index].Tm = sql.NullString{Valid: false}
				data[index].Aid = sql.NullString{Valid: false}
				data[index].Ttno = sql.NullString{Valid: false}
			} else {
				data[index].Tm = sql.NullString{String: "tm text", Valid: true}
				data[index].Aid = sql.NullString{String: "aid text", Valid: true}
				data[index].Ttno = sql.NullString{String: "ttno text", Valid: true}
			}
			data[index].Sn = "sn text"
			data[index].Cus = "cus text"
			data[index].Tr = "tr text"
			data[index].Pid = "pid text"
			data[index].Code = "code text"
			data[index].UpdateTime = time.Now()
		}
		_, err := db.Exec(sqlText, data)
		if err != nil {
			return err
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
	err = merge(db)
	if err != nil {
		t.Error(err)
		return
	}
}
