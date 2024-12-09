package main

import (
	"database/sql"
	"fmt"
	"os"

	go_ora "github.com/sijms/go-ora/v2"
)

type Product struct {
	Id   int    `udt:"ID"`
	Name string `udt:"NAME"`
	Desc string `udt:"DESCR"`
}
type Customer struct {
	Id         int       `udt:"ID"`
	Name       string    `udt:"NAME"`
	Products   []Product `udt:"PRODUCTS"`
	FavProduct *Product  `udt:"FAV_PRODUCT"`
}

func execCmd(db *sql.DB, stmts ...string) error {
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			if len(stmts) > 1 {
				return fmt.Errorf("error: %v in execuation of stmt: %s", err, stmt)
			} else {
				return err
			}
		}
	}
	return nil
}

func createTypes(db *sql.DB) error {
	return execCmd(db, `
create or replace type PRODUCT as object(
ID number,
NAME varchar2(100),
DESCR varchar2(100)
)`, `create or replace type ProductCol as table of product`, `
create or replace type CUSTOMER as object(
ID number,
FAV_PRODUCT PRODUCT,
NAME varchar2(100),
PRODUCTS ProductCol
)`)
}

func dropTypes(db *sql.DB) error {
	return execCmd(db, "DROP TYPE CUSTOMER", "DROP TYPE ProductCol", "DROP TYPE PRODUCT")
}

func fullInputFullOutput(db *sql.DB) error {
	input := Customer{
		Id:   1,
		Name: "customer_",
		FavProduct: &Product{
			1, "product_1", "description of product_1",
		},
		Products: []Product{
			{1, "prduct_1", "description of product_1"},
			{2, "prduct_2", "description of product_2"},
		},
	}
	var output Customer
	_, err := db.Exec(`
DECLARE
	l_customer customer;
BEGIN
	l_customer := :1;
	l_customer.id := l_customer.id + 2;
	l_customer.name := l_customer.name || '3';
	l_customer.FAV_PRODUCT := PRODUCT(3, 'product_3', 'description of product_3');
	for x in 3..5 loop
		l_customer.products.extend;
		l_customer.products(x) := product(x, 'product_' || x, 'description of product_' || x);
	end loop;
	:2 := l_customer;
END;`, go_ora.Object{Name: "Customer", Value: input},
		go_ora.Object{Name: "Customer", Value: &output})
	if err != nil {
		return err
	}
	fmt.Println("full input full output: ", output)
	if output.FavProduct != nil {
		fmt.Println("fav product: ", output.FavProduct)
	}
	return nil
}

func fullInputNullOutput(db *sql.DB) error {
	input := Customer{
		Id:   1,
		Name: "customer_",
		FavProduct: &Product{
			1, "product_1", "description of product_1",
		},
		Products: []Product{
			{1, "prduct_1", "description of product_1"},
			{2, "prduct_2", "description of product_2"},
		},
	}
	var output Customer
	_, err := db.Exec(`
DECLARE
	l_customer customer;
BEGIN
	l_customer := :1;
	l_customer.id := l_customer.id + 2;
	l_customer.name := l_customer.name || '3';
	l_customer.FAV_PRODUCT := null;
	l_customer.products := null;
	:2 := l_customer;
END;`, go_ora.Object{Name: "Customer", Value: input},
		go_ora.Object{Name: "Customer", Value: &output})
	if err != nil {
		return err
	}
	fmt.Println("full input null output: ", output)
	if output.FavProduct != nil {
		fmt.Println("fav product: ", output.FavProduct)
	}
	return nil
}

func nullInputFullOutput(db *sql.DB) error {
	input := Customer{
		Id:   1,
		Name: "customer_",
	}
	var output Customer
	_, err := db.Exec(`
DECLARE
	l_customer customer;
BEGIN
	l_customer := :1;
	l_customer.id := l_customer.id + 2;
	l_customer.name := l_customer.name || '3';
	l_customer.FAV_PRODUCT := PRODUCT(1, 'product_1', 'description of product_1');
	l_customer.Products := productCol();
	for x in 1..5 loop
		l_customer.products.extend;
		l_customer.products(x) := product(x, 'product_' || x, 'description of product_' || x);
	end loop;
	:2 := l_customer;
END;`, go_ora.Object{Name: "Customer", Value: input},
		go_ora.Object{Name: "customer", Value: &output})
	if err != nil {
		return err
	}
	fmt.Println("null input full output: ", output)
	if output.FavProduct != nil {
		fmt.Println("fav product: ", output.FavProduct)
	}
	return nil
}

func nullObject(db *sql.DB) error {
	var output *Product
	var output2 string
	_, err := db.Exec(`
DECLARE
	l_product Product := null;
BEGIN
	:1 := l_product;
	:2 := 'this is a test';
END;`, go_ora.Object{Name: "Product", Value: &output}, go_ora.Out{Dest: &output2, Size: 20})
	fmt.Println(output)
	fmt.Println(output2)
	return err
}

func nullArray(db *sql.DB) error {
	var output []Product
	// output = append(output, Product{3, "product_3", "desc_3"})
	var output2 string
	_, err := db.Exec(`
DECLARE
	l_products productCol := null;
BEGIN
	:1 := l_products;
	:2 := 'this is a test';
END;`, go_ora.Object{Name: "ProductCol", Value: &output}, go_ora.Out{Dest: &output2, Size: 30})
	fmt.Println(output)
	fmt.Println(output2)
	return err
}

// passing nil nested type and receiving it
// passing nil nested array and receving it
func nullInputNullOutput(db *sql.DB) error {
	input := Customer{
		Id:   1,
		Name: "customer_",
	}
	var output Customer
	_, err := db.Exec(`
DECLARE
	l_customer customer;
BEGIN
	l_customer := :1;
	l_customer.id := l_customer.id + 2;
	l_customer.name := l_customer.name || '3';
	:2 := l_customer;
END;`, go_ora.Object{Name: "Customer", Value: input},
		go_ora.Object{Name: "customer", Value: &output})
	if err != nil {
		return err
	}
	fmt.Println("null input null output: ", output)
	return nil
}

func main() {
	db, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't open database: ", err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("can't close database: ", err)
		}
	}()
	err = createTypes(db)
	if err != nil {
		fmt.Println("can't create types: ", err)
		return
	}
	defer func() {
		err = dropTypes(db)
		if err != nil {
			fmt.Println("can't drop types: ", err)
		}
	}()
	err = go_ora.RegisterType(db, "product", "ProductCol", Product{})
	if err != nil {
		fmt.Println("can't register product: ", err)
		return
	}
	err = go_ora.RegisterType(db, "customer", "", Customer{})
	if err != nil {
		fmt.Println("can't register customer: ", err)
		return
	}
	err = nullInputNullOutput(db)
	if err != nil {
		fmt.Println("can't send null input null output: ", err)
		return
	}
	fmt.Println()
	err = nullInputFullOutput(db)
	if err != nil {
		fmt.Println("can't send null input full output: ", err)
		return
	}
	fmt.Println()
	err = fullInputFullOutput(db)
	if err != nil {
		fmt.Println("can't send full input full output: ", err)
		return
	}
	fmt.Println()
	err = fullInputNullOutput(db)
	if err != nil {
		fmt.Println("can't send full input null output: ", err)
		return
	}
	fmt.Println()
	err = nullArray(db)
	if err != nil {
		fmt.Println("can't get null array: ", err)
		return
	}
	fmt.Println()
	err = nullObject(db)
	if err != nil {
		fmt.Println("can't get null object: ", err)
		return
	}
}
