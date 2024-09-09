package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
)

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
create or replace type childType2 as object(
	ID number,
	NAME varchar2(100),
	LDATE DATE
)`, `create or replace type childType2Col as table of childType2`, `
create or replace type childType as object(
	ID number,
	NAME varchar2(100),
	DATA childType2Col,
	child childType2
)`, `create or replace type childTypeCol as table of childType`, `
create or replace type parentType as object(
	ID number,
	VALUE varchar2(100),
	DATA childTypeCol,
	child childType
)`, `create or replace type parentTypeCol as table of parentType`)
}

func dropTypes(db *sql.DB) error {
	return execCmd(db,
		"DROP TYPE parentTypeCol",
		"DROP TYPE parentType",
		"DROP TYPE childTypeCol",
		"DROP TYPE childType",
		"DROP TYPE childType2Col",
		"DROP TYPE childType2")
}

type typeChild2 struct {
	Id   int       `udt:"ID"`
	Name string    `udt:"NAME"`
	Date time.Time `udt:"LDATE"`
}
type typeChild struct {
	Id    int          `udt:"ID"`
	Name  string       `udt:"NAME"`
	Data  []typeChild2 `udt:"DATA"`
	Child typeChild2   `udt:"CHILD"`
}

type typeParent struct {
	Id    int         `udt:"ID"`
	Value string      `udt:"VALUE"`
	Data  []typeChild `udt:"DATA"`
	Child typeChild   `udt:"CHILD"`
}

var refDate = time.Date(2024, 1, 11, 19, 19, 19, 0, time.UTC)

func createParent(index int, date time.Time) typeParent {
	var parent typeParent
	parent.Id = index + 1
	parent.Value = fmt.Sprintf("parent_%d", index+1)
	for x := 0; x < 3; x++ {
		temp := typeChild{}
		temp.Id = x + 1
		temp.Name = fmt.Sprintf("child_%d", x+1)
		for i := 0; i < 5; i++ {
			temp.Data = append(temp.Data, typeChild2{Id: i + 1, Name: fmt.Sprintf("child2_%d", i+1), Date: date})
		}
		temp.Child = typeChild2{6, "child2_6", date}
		parent.Data = append(parent.Data, temp)
	}
	parent.Child = typeChild{Id: 4, Name: "child_4"}
	for i := 0; i < 5; i++ {
		parent.Child.Data = append(parent.Child.Data, typeChild2{Id: i + 1, Name: fmt.Sprintf("child2_%d", i+1), Date: date})
	}
	parent.Child.Child = typeChild2{6, "child2_6", date}
	return parent
}

func outputPar(db *sql.DB) error {
	parent := typeParent{}
	_, err := db.Exec(`
DECLARE
	v_child2 childType2;
	v_child childType;
	v_parent parentType := parentType(null,null,null, null);
BEGIN
	v_parent.id := 1;
	v_parent.value := 'parent';
	v_parent.DATA := childTypeCol();
	for i in 1..3 loop
		v_child := childType(i,'child_' || i, childType2Col(), null);
		for x in 1..5 loop
			v_child2 := childType2(x, 'child2_' || x, :ldate);
			v_child.DATA.extend;
			v_child.DATA(x) := v_child2;
		end loop;
		v_child.child := childType2(6, 'child2_6', :ldate);
		v_parent.DATA.extend;
		v_parent.DATA(i) := v_child;
	end loop;
	v_child := childType(4,'child_4', childType2Col(), null);
	for x in 1..5 loop
		v_child2 := childType2(x, 'child2_' || x, :ldate);
		v_child.DATA.extend;
		v_child.DATA(x) := v_child2;
	end loop;
	v_child.child := childType2(6, 'child2_6', :ldate);
	v_parent.CHILD := null;
	:output := v_parent;
END;`, sql.Named("ldate", refDate), sql.Named("output", go_ora.Out{Dest: &parent}))
	fmt.Println("output parameter example: ")
	fmt.Println(parent)
	return err
}

func inputPar(db *sql.DB) error {
	parent := createParent(0, refDate)
	var output typeParent
	_, err := db.Exec(`
DECLARE
	parent parentType;
	v_child childType;
	v_child2 childType2;
begin
	parent := :parent;
	parent.id := parent.id + :id;
	parent.value := parent.value || :name;
	for i in 1..parent.data.count loop
		v_child := parent.data(i);
		v_child.id := v_child.id + :id;
		v_child.name := v_child.name || :name;
		for x in 1..v_child.data.count loop
			v_child2 := v_child.data(x);
			v_child2.id := v_child2.id + :id;
			v_child2.name := v_child2.name || :name;
			v_child2.ldate := :ldate;
			v_child.data(x) := v_child2;
		end loop;
		v_child2 := v_child.child;
		v_child2.id := v_child2.id + :id;
		v_child2.name := v_child2.name || :name;
		v_child2.ldate := :ldate;
		v_child.child := v_child2;
		parent.data(i) := v_child;
	end loop;
	v_child := parent.child;
	v_child.id := v_child.id + :id;
	v_child.name := v_child.name || :name;
	for x in 1..v_child.data.count loop
		v_child2 := v_child.data(x);
		v_child2.id := v_child2.id + :id;
		v_child2.name := v_child2.name || :name;
		v_child2.ldate := :ldate;
		v_child.data(x) := v_child2;
	end loop;
	v_child2 := v_child.child;
	v_child2.id := v_child2.id + :id;
	v_child2.name := v_child2.name || :name;
	v_child2.ldate := :ldate;
	v_child.child := v_child2;
	parent.child := v_child;
	:output := parent;
end;`, sql.Named("parent", parent), sql.Named("id", 5), sql.Named("name", "_verified"),
		sql.Named("ldate", refDate), sql.Named("output", go_ora.Out{Dest: &output}))
	if err != nil {
		return err
	}
	fmt.Println("input parameters example")
	fmt.Println(output)
	return nil
}

func outputParArray(db *sql.DB) error {
	var parents []typeParent
	_, err := db.Exec(`
DECLARE
	v_child2 childType2;
	v_child childType;
	v_parent parentType := parentType(null,null,null, null);
	v_parents parentTypeCol;
BEGIN
	v_parent.id := 1;
	v_parent.value := 'parent';
	v_parent.DATA := childTypeCol();
	for i in 1..3 loop
		v_child := childType(i,'child_' || i, childType2Col(), null);
		for x in 1..5 loop
			v_child2 := childType2(x, 'child2_' || x, :ldate);
			v_child.DATA.extend;
			v_child.DATA(x) := v_child2;
		end loop;
		v_child.child := childType2(6, 'child2_6', :ldate);
		v_parent.DATA.extend;
		v_parent.DATA(i) := v_child;
	end loop;
	v_child := childType(4,'child_4', childType2Col(), null);
	for x in 1..5 loop
		v_child2 := childType2(x, 'child2_' || x, :ldate);
		v_child.DATA.extend;
		v_child.DATA(x) := v_child2;
	end loop;
	v_child.child := childType2(6, 'child2_6', :ldate);
	v_parent.CHILD := v_child;
	v_parents := parentTypeCol();
	for y in 1..3 loop
		v_parents.extend;
		v_parents(y) := v_parent;
	end loop;
	:output := v_parents;
END;`, sql.Named("ldate", refDate), sql.Named("output", go_ora.Out{Dest: &parents, Size: 10}))
	fmt.Println("output parameter array example: ")
	fmt.Println(parents)
	return err
}

func inputParArray(db *sql.DB) error {
	var parents []typeParent
	parents = append(parents, createParent(0, refDate),
		createParent(1, refDate),
		createParent(2, refDate))
	var output typeParent
	_, err := db.Exec(`
DECLARE
	parents parentTypeCol;
	parent parentType;
	v_child childType;
	v_child2 childType2;
begin
	parents := :parents;
	parent := parents(1);
	parent.id := parent.id + :id;
	parent.value := parent.value || :name;
	for i in 1..parent.data.count loop
		v_child := parent.data(i);
		v_child.id := v_child.id + :id;
		v_child.name := v_child.name || :name;
		for x in 1..v_child.data.count loop
			v_child2 := v_child.data(x);
			v_child2.id := v_child2.id + :id;
			v_child2.name := v_child2.name || :name;
			v_child2.ldate := :ldate;
			v_child.data(x) := v_child2;
		end loop;
		v_child2 := v_child.child;
		v_child2.id := v_child2.id + :id;
		v_child2.name := v_child2.name || :name;
		v_child2.ldate := :ldate;
		v_child.child := v_child2;
		parent.data(i) := v_child;
	end loop;
	v_child := parent.child;
	v_child.id := v_child.id + :id;
	v_child.name := v_child.name || :name;
	for x in 1..v_child.data.count loop
		v_child2 := v_child.data(x);
		v_child2.id := v_child2.id + :id;
		v_child2.name := v_child2.name || :name;
		v_child2.ldate := :ldate;
		v_child.data(x) := v_child2;
	end loop;
	v_child2 := v_child.child;
	v_child2.id := v_child2.id + :id;
	v_child2.name := v_child2.name || :name;
	v_child2.ldate := :ldate;
	v_child.child := v_child2;
	parent.child := v_child;
	:output := parent;
end;`, sql.Named("parents", parents), sql.Named("id", 5), sql.Named("name", "_verified"),
		sql.Named("ldate", refDate), sql.Named("output", go_ora.Out{Dest: &output}))
	if err != nil {
		return err
	}
	fmt.Println("input parameter array example: ")
	fmt.Println("output: ", output)
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
	err = go_ora.RegisterType(db, "childType2", "childType2Col", typeChild2{})
	if err != nil {
		fmt.Println("can't register child type2: ", err)
		return
	}
	err = go_ora.RegisterType(db, "childType", "childTypeCol", typeChild{})
	if err != nil {
		fmt.Println("can't register child type: ", err)
		return
	}
	err = go_ora.RegisterType(db, "parentType", "parentTypeCol", typeParent{})
	if err != nil {
		fmt.Println("can't register parent type: ", err)
		return
	}
	err = inputPar(db)
	if err != nil {
		fmt.Println("can't input par: ", err)
		return
	}
	fmt.Println()
	err = outputPar(db)
	if err != nil {
		fmt.Println("can't output par: ", err)
		return
	}
	fmt.Println()
	err = inputParArray(db)
	if err != nil {
		fmt.Println("can't input par array: ", err)
		return
	}
	fmt.Println()
	err = outputParArray(db)
	if err != nil {
		fmt.Println("can't output par array: ", err)
		return
	}
}
