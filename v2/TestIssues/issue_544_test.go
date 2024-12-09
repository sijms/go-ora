// insert and query json data
// require oracle 21c
package TestIssues

//
//import (
//	"database/sql"
//	"testing"
//)
//
//func TestIssue544(t *testing.T) {
//	var createInsert = func(db *sql.DB) error {
//		return execCmd(db, `CREATE TABLE TTB_544(
//			id NUMBER PRIMARY KEY,
//			data JSON
//		)`, `INSERT INTO TTB_544(ID, DATA) VALUES (1, '{"name": "John", "age": 30}')`)
//	}
//	var dropTable = func(db *sql.DB) error {
//		return execCmd(db, `DROP TABLE TTB_544 PURGE`)
//	}
//
//	db, err := getDB()
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	defer func() {
//		err := db.Close()
//		if err != nil {
//			t.Error(err)
//		}
//	}()
//	err = createInsert(db)
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	defer func() {
//		err := dropTable(db)
//		if err != nil {
//			t.Error(err)
//		}
//	}()
//	var (
//		id   int
//		data []byte
//	)
//	err = db.QueryRow("SELECT ID, DATA FROM TTB_544").Scan(&id, &data)
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	if id != 1 {
//		t.Errorf("want 1, got %d", id)
//		return
//	}
//	if string(data) != `{"name":"John","age":30}` {
//		t.Errorf("want %s, got %s", `{"name":"John","age":30}`, string(data))
//		return
//	}
//}
