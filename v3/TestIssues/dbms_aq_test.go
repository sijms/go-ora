package TestIssues

import (
	"bytes"
	"testing"

	"github.com/sijms/go-ora/v3/aq"
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
	err = execCmd(db, `BEGIN
	DBMS_AQADM.CREATE_QUEUE_TABLE(queue_table => 'GO_ORA_QT', queue_payload_type => 'RAW');
	DBMS_AQADM.CREATE_QUEUE(queue_name => 'GO_ORA_QU', queue_table => 'GO_ORA_QT');
	DBMS_AQADM.START_QUEUE(queue_name => 'GO_ORA_QU');
END;`)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = execCmd(db, `BEGIN
	DBMS_AQADM.STOP_QUEUE(queue_name => 'GO_ORA_QU');
	DBMS_AQADM.DROP_QUEUE(queue_name => 'GO_ORA_QU');
	DBMS_AQADM.DROP_QUEUE_TABLE(queue_table => 'GO_ORA_QT');
END;`)
		if err != nil {
			t.Error(err)
		}
	}()
	queue, err := aq.CreateQueue(db, "GO_ORA_QU", aq.RAW, "")
	if err != nil {
		t.Error(err)
		return
	}
	queue.SetAutoCommit(true)
	expected := []byte("this is a test")
	message, err := queue.NewMessage(expected)
	if err != nil {
		t.Error(err)
		return
	}
	err = queue.Enqueue(message)
	if err != nil {
		t.Error(err)
		return
	}
	output, err := queue.Dequeue(aq.DefaultDequeueOptions())
	if err != nil {
		t.Error(err)
		return
	}
	got, ok := output.Payload.([]byte)
	if !ok || !bytes.Equal(got, expected) {
		t.Errorf("expected payload %q and got %#v", expected, output.Payload)
	}
}
