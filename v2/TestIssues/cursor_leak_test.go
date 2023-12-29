package TestIssues

import (
	"testing"
)

func TestCursorLeak(t *testing.T) {
	return
	//var execQuery = func(db *sql.DB, query string) {
	//	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1)*time.Second)
	//	defer cancel()
	//	rows, err := db.QueryContext(ctx, query)
	//	if err != nil {
	//		t.Error(err)
	//		return
	//	}
	//	defer func() {
	//		err = rows.Close()
	//		if err != nil {
	//			t.Error(err)
	//		}
	//	}()
	//}
	//
	//var pingDB = func(db *sql.DB) {
	//	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	//	defer cancel()
	//	err := db.PingContext(ctx)
	//	if err != nil {
	//		t.Error(err)
	//	}
	//}
	//
	//db, err := getDB()
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//defer func() {
	//	err = db.Close()
	//	if err != nil {
	//		t.Error(err)
	//	}
	//}()
	//db.SetMaxOpenConns(8)
	//sqlText := "Select * from dba_objects a, dba_objects b, dba_objects c, dba_objects d, dba_objects e"
	//queries := [4]string{sqlText, sqlText, sqlText, sqlText}
	//scheduler, err := gocron.NewScheduler()
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//for _, query := range queries {
	//	if _, err := scheduler.Every(30).Second().Do(execQuery, db, query); err != nil {
	//		log.Fatalf("Err job: %v", err)
	//	}
	//}
	//if _, err := scheduler.Every(60).Second().Do(pingDB, db); err != nil {
	//	log.Fatalf("Err job: %v", err)
	//}
	//
	//scheduler.StartBlocking()
}
