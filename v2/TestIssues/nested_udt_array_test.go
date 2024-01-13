/*
	file test:

1- nested UDT array [in/out] as object
2- nested UDT array [in/out] as array
3- nested UDT object
data used
typeParent --> typeChild, []typeChild
typeChild --> typeChild2, []typeChild2
also file test nested udt as well
*/
package TestIssues

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"testing"
	"time"
)

func TestNestedUDTArray(t *testing.T) {
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
	var createParent = func(index int, date time.Time) typeParent {
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
	var checkChildType2 = func(input *typeChild2, index, idShift int, extraText string) error {
		if input.Id != index+1+idShift {
			return fmt.Errorf("expected child2.Id: %d and got: %d at index: %d",
				index+1+idShift, input.Id, index)
		}
		if input.Name != fmt.Sprintf("child2_%d%s", index+1, extraText) {
			return fmt.Errorf("expected child2.Name: %s and got: %s at index: %d",
				fmt.Sprintf("child2_%d%s", index+1, extraText), input.Name, index)
		}
		if !isEqualTime(input.Date, refDate) {
			return fmt.Errorf("expected child2.Date: %v and got: %v at index: %d",
				refDate, input.Date, index)
		}
		return nil
	}
	var checkChildType = func(input *typeChild, index, idShift int, extraText string) error {
		for index2, temp := range input.Data {
			err := checkChildType2(&temp, index2, idShift, extraText)
			if err != nil {
				return err
			}
		}
		if input.Id != index+1+idShift {
			return fmt.Errorf("expected child.Id: %d and got: %d at index: %d",
				index+1+idShift, input.Id, index)
		}
		if input.Name != fmt.Sprintf("child_%d%s", index+1, extraText) {
			return fmt.Errorf("expected child.Name: %s and got: %s at index: %d",
				fmt.Sprintf("child_%d%s", index+1, extraText), input.Name, index)
		}
		return checkChildType2(&input.Child, 5, idShift, extraText)
	}
	var checkParentType = func(input *typeParent, index, idShift int, extraText string) error {
		for index2, temp := range input.Data {
			err := checkChildType(&temp, index2, idShift, extraText)
			if err != nil {
				return err
			}
		}
		if input.Id != index+1+idShift {
			return fmt.Errorf("expected parent.Id: %d and got: %d at index: %d", index+1+idShift, input.Id, index)
		}
		if input.Value != fmt.Sprintf("parent_%d%s", index+1, extraText) {
			return fmt.Errorf("expected parent.Value: %s and got: %s at index: %d",
				fmt.Sprintf("parent_%d%s", index+1, extraText), input.Value, index)
		}
		return checkChildType(&input.Child, 3, idShift, extraText)
	}
	var createTypes = func(db *sql.DB) error {
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
	var dropTypes = func(db *sql.DB) error {
		return execCmd(db,
			"DROP TYPE parentTypeCol",
			"DROP TYPE parentType",
			"DROP TYPE childTypeCol",
			"DROP TYPE childType",
			"DROP TYPE childType2Col",
			"DROP TYPE childType2")
	}
	var outputPar = func(db *sql.DB) error {
		var parent = typeParent{}
		_, err := db.Exec(`
DECLARE
	v_child2 childType2;
	v_child childType;
	v_parent parentType := parentType(null,null,null, null);
BEGIN
	v_parent.id := 1;
	v_parent.value := 'parent_1';
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
	:output := v_parent;
END;`, sql.Named("ldate", refDate), sql.Named("output", go_ora.Out{Dest: &parent}))
		if err != nil {
			return err
		}
		return checkParentType(&parent, 0, 0, "")
	}

	var inputPar = func(db *sql.DB) error {
		var input = time.Date(2024, 1, 11, 19, 19, 19, 0, time.UTC)
		var parent = createParent(0, time.Now())
		var output typeParent
		var id = 5
		var name = "_verified"
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
end;`, sql.Named("parent", parent), sql.Named("id", id), sql.Named("name", name),
			sql.Named("ldate", input), sql.Named("output", go_ora.Out{Dest: &output}))
		if err != nil {
			return err
		}
		return checkParentType(&output, 0, id, name)
	}

	var inputParArray = func(db *sql.DB) error {
		var parents []typeParent
		parents = append(parents, createParent(0, refDate),
			createParent(1, refDate),
			createParent(2, refDate))
		var output typeParent
		id := 5
		name := "_verified"
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
	end;`, sql.Named("parents", parents), sql.Named("id", id), sql.Named("name", name),
			sql.Named("ldate", refDate), sql.Named("output", go_ora.Out{Dest: &output}))
		if err != nil {
			return err
		}
		return checkParentType(&output, 0, id, name)
	}

	var outputParArray = func(db *sql.DB) error {
		var parents []typeParent
		_, err := db.Exec(`
	DECLARE
		v_child2 childType2;
		v_child childType;
		v_parent parentType := parentType(null,null,null, null);
		v_parents parentTypeCol;
	BEGIN
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
			v_parent.value := 'parent_' || y;
			v_parent.id := y;
			v_parents(y) := v_parent;
		end loop;
		:output := v_parents;
	END;`, sql.Named("ldate", refDate), sql.Named("output", go_ora.Out{Dest: &parents, Size: 10}))
		if err != nil {
			return err
		}
		if len(parents) != 3 {
			return fmt.Errorf("expected output length: %d and got %d", 3, len(parents))
		}
		for index, parent := range parents {
			err = checkParentType(&parent, index, 0, "")
			if err != nil {
				return err
			}
		}
		return nil
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
	err = go_ora.RegisterType(db, "childType2", "childType2Col", typeChild2{})
	if err != nil {
		t.Error(err)
		return
	}
	err = go_ora.RegisterType(db, "childType", "childTypeCol", typeChild{})
	if err != nil {
		t.Error(err)
		return
	}
	err = go_ora.RegisterType(db, "parentType", "parentTypeCol", typeParent{})
	if err != nil {
		t.Error(err)
		return
	}
	err = outputPar(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = inputPar(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = inputParArray(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = outputParArray(db)
	if err != nil {
		t.Error(err)
		return
	}
}
