package dbms

import (
	"database/sql"
	"database/sql/driver"
	"errors"

	go_ora "github.com/sijms/go-ora/v2"
)

type AQ struct {
	conn          *sql.DB
	Name          string `db:"QUEUE_NAME"`
	TableName     string `db:"TB_NAME"`
	TypeName      string `db:"TYPE_NAME"`
	Owner         string
	MaxRetry      int64  `db:"MAX_RETRY"`
	RetryDelay    int64  `db:"RETRY_DELAY"`
	RetentionTime int64  `db:"RETENTION_TIME"`
	Comment       string `db:"COMMENT"`
}

func NewAQ(conn *sql.DB, name, typeName string) *AQ {
	output := &AQ{
		conn:          conn,
		Name:          name,
		TableName:     name + "_TB",
		TypeName:      typeName,
		RetentionTime: -1,
		Comment:       name,
	}
	return output
}

func (aq *AQ) validate() error {
	if aq.conn == nil {
		return errors.New("no connection defined for AQ type")
	}
	if len(aq.Name) == 0 {
		return errors.New("queue name cannot be null")
	}
	if len(aq.TypeName) == 0 {
		return errors.New("type name cannot be null")
	}
	return nil
}

func (aq *AQ) Create() error {
	err := aq.validate()
	if err != nil {
		return err
	}
	sqlText := `BEGIN
	DBMS_AQADM.CREATE_QUEUE_TABLE (:TB_NAME, :TYPE_NAME);

	DBMS_AQADM.CREATE_QUEUE (
		:QUEUE_NAME, :TB_NAME, DBMS_AQADM.NORMAL_QUEUE, 
		:MAX_RETRY, :RETRY_DELAY, :RETENTION_TIME, 
		FALSE, :COMMENT);
END;`
	_, err = aq.conn.Exec(sqlText, aq)
	return err
}

func (aq *AQ) Drop() error {
	err := aq.validate()
	if err != nil {
		return err
	}
	sqlText := `BEGIN
	DBMS_AQADM.DROP_QUEUE(:QUEUE_NAME, FALSE);
	DBMS_AQADM.DROP_QUEUE_TABLE(:TB_NAME);
END;`
	_, err = aq.conn.Exec(sqlText, aq) // sql.Named("QUEUE_NAME", aq.Name),
	// sql.Named("TABLE_NAME", aq.TableName))
	return err
}

// enable both enqueue and dequeue
func (aq *AQ) Start(enqueue, dequeue bool) error {
	err := aq.validate()
	if err != nil {
		return err
	}
	_, err = aq.conn.Exec(`BEGIN
dbms_aqadm.start_queue (queue_name => :QUEUE_NAME, 
                       enqueue => :ENQUEUE , 
                       dequeue => :DEQUEUE); 
 END;`, aq.Name, go_ora.PLBool(enqueue), go_ora.PLBool(dequeue))
	return err
}

// disable both enqueue and dequeue
func (aq *AQ) Stop(enqueue, dequeue bool) error {
	err := aq.validate()
	if err != nil {
		return err
	}
	_, err = aq.conn.Exec(`BEGIN
dbms_aqadm.stop_queue(queue_name => :QUEUE_NAME, 
                       enqueue => :ENQUEUE , 
                       dequeue => :DEQUEUE); 
 END;`, aq.Name, go_ora.PLBool(enqueue), go_ora.PLBool(dequeue))
	return err
}

func (aq *AQ) Dequeue(message driver.Value, messageSize int) (messageID []byte, err error) {
	err = aq.validate()
	if err != nil {
		return
	}
	sqlText := `DECLARE
	dequeue_options dbms_aq.dequeue_options_t;
	message_properties	dbms_aq.message_properties_t;
BEGIN
	dequeue_options.VISIBILITY := DBMS_AQ.IMMEDIATE;
	DBMS_AQ.DEQUEUE (
		queue_name => :QUEUE_NAME,
		dequeue_options => dequeue_options,
		message_properties => message_properties,
		payload => :MSG,
		msgid => :MSG_ID);
END;`
	_, err = aq.conn.Exec(sqlText, sql.Named("QUEUE_NAME", aq.Name),
		sql.Named("MSG", go_ora.Out{Dest: message, Size: messageSize}),
		sql.Named("MSG_ID", go_ora.Out{Dest: &messageID, Size: 100}))
	return
}

func (aq *AQ) Enqueue(message driver.Value) (messageID []byte, err error) {
	err = aq.validate()
	if err != nil {
		return
	}
	sqlText := `DECLARE
	enqueue_options dbms_aq.enqueue_options_t;
	message_properties dbms_aq.message_properties_t;
BEGIN
	DBMS_AQ.ENQUEUE (
	  queue_name => :QUEUE_NAME,
	  enqueue_options => enqueue_options,
	  message_properties => message_properties,
	  payload => :MSG,
	  msgid => :MSG_ID);
END;`
	_, err = aq.conn.Exec(sqlText, sql.Named("QUEUE_NAME", aq.Name),
		sql.Named("MSG", message),
		sql.Named("MSG_ID", go_ora.Out{Dest: &messageID, Size: 100}))
	return
}
