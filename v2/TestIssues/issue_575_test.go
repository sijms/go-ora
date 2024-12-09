package TestIssues

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"
)

func TestIssue575(t *testing.T) {
	var create = func(db *sql.DB) error {
		return execCmd(db, `CREATE TABLE TTB_575 (
	RN NUMBER,
	ID_ NVARCHAR2(64) NOT NULL,
	REV_ INTEGER,
	NAME_ NVARCHAR2(255),
	DEPLOYMENT_ID_ NVARCHAR2(64),
	BYTES_ BLOB,
	STR_ CLOB,
	GENERATED_ NUMBER(1)
)`)
	}
	var drop = func(db *sql.DB) error {
		return execCmd(db, `DROP TABLE TTB_575 PURGE`)
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
	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		t.Error(err)
		return
	}
	var stmt *sql.Stmt
	stmt, err = tx.Prepare(`INSERT INTO TTB_575(RN, ID_, REV_, NAME_, DEPLOYMENT_ID_, BYTES_, STR_, GENERATED_) 
VALUES (:1, :2, :3, :4, :5, :6, :7, :8)`)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	largeBytes := bytes.Repeat([]byte{0x3}, 0xFFFF)
	smallBytes := bytes.Repeat([]byte{0x2}, 0xFF)
	largeString := strings.Repeat("a", 0xFFFF)
	smallString := strings.Repeat("b", 0xFF)

	// insert large bytes and large string
	_, err = stmt.Exec("100",
		"1701058348821917698",
		"1",
		"信息服务接入表.bpmn_64fa805addca7e2d490debe8.png",
		"1701058347710427137",
		largeBytes,
		largeString,
		nil)
	if err != nil {
		t.Error(err)
		err = tx.Rollback()
		if err != nil {
			t.Error(err)
		}
		return
	}
	// insert small bytes and small string
	_, err = stmt.Exec("101",
		"1701078318708568066",
		"1",
		"var-Activity_1e3bcfyAssigneeList",
		nil,
		smallBytes,
		smallString,
		nil)
	if err != nil {
		t.Error(err)
		err = tx.Rollback()
		if err != nil {
			t.Error(err)
		}
		return
	}
	// insert large bytes and small string
	_, err = stmt.Exec("102",
		"1701058348821917698",
		"1",
		"信息服务接入表.bpmn_64fa805addca7e2d490debe8.png",
		"1701058347710427137",
		largeBytes,
		smallString,
		nil)
	if err != nil {
		t.Error(err)
		err = tx.Rollback()
		if err != nil {
			t.Error(err)
		}
		return
	}
	// insert small bytes and large string
	_, err = stmt.Exec("103",
		"1701058348821917698",
		"1",
		"信息服务接入表.bpmn_64fa805addca7e2d490debe8.png",
		"1701058347710427137",
		smallBytes,
		largeString,
		nil)
	if err != nil {
		t.Error(err)
		err = tx.Rollback()
		if err != nil {
			t.Error(err)
		}
		return
	}
	err = tx.Commit()
	if err != nil {
		t.Error(err)
		return
	}
}
