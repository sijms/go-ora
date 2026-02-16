package TestIssues

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v3"
	"testing"
	"time"
)

func TestIssue482(t *testing.T) {
	type test1 struct {
		Name string `udt:"name"`
	}
	var createTypes = func(db *sql.DB) error {
		return execCmd(db, `CREATE TYPE test1 AS OBJECT ( name varchar2(256) )`,
			`CREATE OR REPLACE TYPE test1collection AS TABLE OF test1`)
	}
	var dropTypes = func(db *sql.DB) error {
		return execCmd(db, `DROP TYPE test1collection`, `DROP TYPE test1`)
	}
	var setupQueue = func(db *sql.DB) error {
		return execCmd(db,
			`BEGIN DBMS_AQADM.CREATE_QUEUE_TABLE ( queue_table => 'test1table', queue_payload_type => 'test1' ); END;`,
			`BEGIN DBMS_AQADM.CREATE_QUEUE ( queue_name => 'test1queue', queue_table => 'test1table' ); END;`,
			`BEGIN DBMS_AQADM.START_QUEUE ( queue_name => 'test1queue', enqueue => TRUE ); END;`)
	}
	var stopQueue = func(db *sql.DB) error {
		return execCmd(db,
			`BEGIN DBMS_AQADM.STOP_QUEUE(queue_name => 'test1queue'); END;`,
			`BEGIN DBMS_AQADM.DROP_QUEUE(queue_name => 'test1queue'); END;`,
			`BEGIN DBMS_AQADM.DROP_QUEUE_TABLE(queue_table => 'test1table'); END;`)
	}
	var enqueue = func(db *sql.DB, message test1) error {
		_, err := db.Exec(`
DECLARE
	enqueueOptions			DBMS_AQ.enqueue_options_t;
	messageProperties		DBMS_AQ.message_properties_t;
	msgID_raw				RAW(100);
BEGIN
	DBMS_AQ.ENQUEUE(
		queue_name => 'test1queue',
		enqueue_options => enqueueOptions,
		message_properties => messageProperties,
		payload => :1,
		msgid => msgID_raw
	);
END;`, message)
		return err
	}

	var dequeueArray = func(db *sql.DB, arraySize int, waitTime time.Duration) ([]test1, error) {
		sqlText := fmt.Sprintf(`
DECLARE
	dequeueOptions			DBMS_AQ.dequeue_options_t;
	messagePropertiesArray	DBMS_AQ.message_properties_array_t;
	msgIDArray				DBMS_AQ.msgid_array_t;
BEGIN
	dequeueOptions.WAIT := %d;
	:1 := DBMS_AQ.DEQUEUE_ARRAY(
		queue_name => 'test1queue',
		dequeue_options => dequeueOptions,
		array_size => %d,
		message_properties_array => messagePropertiesArray,
		payload_array => :2,
		msgid_array => msgIDArray
	);
END;`, (waitTime / time.Second), arraySize)
		var nMessages sql.NullInt64
		var messages []test1
		_, err := db.Exec(sqlText, go_ora.Out{Dest: &nMessages}, go_ora.Out{Dest: &messages})
		return messages, err
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
	err = createTypes(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropTypes(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = go_ora.RegisterType(db, "test1", "test1collection", test1{})
	if err != nil {
		t.Error(err)
		return
	}
	err = setupQueue(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = stopQueue(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = enqueue(db, test1{Name: "Test 1"})
	if err != nil {
		t.Error(err)
		return
	}
	messages, err := dequeueArray(db, 2, 5*time.Second)
	if err != nil {
		t.Error(err)
		return
	}
	if len(messages) != 1 {
		t.Errorf("expected message count: %d and received count: %d", 1, len(messages))
		return
	}
	if messages[0].Name != "Test 1" {
		t.Errorf("expected message: %s and received message: %s", "Test 1", messages[0].Name)
	}
}
