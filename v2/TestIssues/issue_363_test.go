package TestIssues

import (
	go_ora "github.com/sijms/go-ora/v2"
	"testing"
	"xorm.io/xorm"
)

type myStruct2 struct {
	ID int64  `xorm:"pk autoincr <- 'id'"`
	T  string `xorm:"string"`
}

func (myStruct2) TableName() string { return "ISSUE_363" }

func TestIssue363(t *testing.T) {
	url := go_ora.BuildUrl(server, port, service, username, password, urlOptions)
	engine, err := xorm.NewEngine("oracle", url)
	if err != nil {
		t.Error(err)
		return
	}
	if err := engine.CreateTables(&myStruct2{}); err != nil {
		t.Error(err)
		return
	}
	defer func() {
		_, err = engine.Exec("DROP TABLE ISSUE_363 PURGE")
		if err != nil {
			t.Error(err)
		}
		_, err = engine.Exec("DROP SEQUENCE SEQ_ISSUE_363")
		if err != nil {
			t.Error(err)
		}
	}()
	test2 := myStruct2{ID: 1, T: "str2"}
	test3 := myStruct2{ID: 2, T: "str2"}

	_, err = engine.Insert(test2, test3)
	if err != nil {
		t.Error(err)
		return
	}
	var result []myStruct2
	query := engine.NoAutoCondition().Table(myStruct2{}).Where(`"string"='str2' AND "id"<>0`).ForUpdate()
	err = query.Find(&result)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(result)
}
