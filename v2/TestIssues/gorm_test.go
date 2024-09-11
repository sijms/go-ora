package TestIssues

import (
	"testing"
	"time"

	oracle "github.com/godoes/gorm-oracle"
	go_ora "github.com/sijms/go-ora/v2"
	"gorm.io/gorm"
)

func TestGorm(t *testing.T) {
	type Product struct {
		gorm.Model
		Code  string
		Price uint
	}

	db, err := gorm.Open(oracle.Open(go_ora.BuildUrl(server, port, service, username, password, urlOptions)), &gorm.Config{})
	if err != nil {
		t.Error(err)
		return
	}
	err = db.AutoMigrate(&Product{})
	if err != nil {
		t.Error(err)
		return
	}
	// Create
	db.Create(&Product{Model: gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}, Code: "D42", Price: 100})

	// drop
	defer func() {
		db.Exec("drop table products purge")
	}()
	// Read
	var product Product
	db.First(&product, 1) // find product with primary key = 1
	t.Log(product)
	db.First(&product, "code = ?", "D42") // find product with code = D42
	t.Log(product)
}
