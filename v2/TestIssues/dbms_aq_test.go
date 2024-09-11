package TestIssues

/*
import (
	"github.com/sijms/go-ora/dbms"
	"testing"
)

func TestDBMSAQ(t *testing.T) {
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
	aq := dbms.NewAQ(db, "GO_ORA_QU", "RAW")
	err = aq.Create()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = aq.Drop()
		if err != nil {
			t.Error(err)
		}
	}()
	err = aq.Start(true, true)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = aq.Stop(true, true)
		if err != nil {
			t.Error(err)
		}
	}()
	messageID, err := aq.Enqueue([]byte("this is a test"))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Enqueue message ID: ", messageID)
	var output []byte
	messageID, err = aq.Dequeue(&output, 100)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Dequeue message ID: ", messageID)
	t.Log("message: ", string(output))
}
*/
