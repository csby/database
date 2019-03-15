package mysql

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	entity := &entity{}
	err := entity.Parse(nil)
	if err == nil {
		t.Error("empty struct should be error")
	}

	entityBase := &TabEntityBase{}
	err = entity.Parse(entityBase)
	if err == nil {
		t.Error("empty struct should be error")
	}

	entity1 := &TabEntity1{
		UserID:   1,
		UserName: "Name 1",
	}
	err = entity.Parse(entity1)
	if err != nil {
		t.Error(err)
	}
	if entity.name != "`tabTest`" {
		t.Error("table name error")
	}
	if len(entity.fields) != 4 {
		t.Fatal("field count error: expect=4, actual=", len(entity.fields))
	}

	entity2 := &TabEntity2{
		UserID:   2,
		UserName: "Name 2",
	}
	err = entity.Parse(entity2)
	if err != nil {
		t.Error(err)
	}
	if entity.name != "`tabTest2`" {
		t.Error("table name error: expect=`tabTest2`, actual=", entity.name)
	}
	checkField(t, entity.fieldByName("`field22`"), "`field22`", "=", entity2.Field22, &entity2.Field22)
	checkField(t, entity.fieldByName("`userId`"), "`userId`", "in", entity2.UserID, &entity2.UserID)
	checkField(t, entity.fieldByName("`userName`"), "`userName`", "like", entity2.UserName, &entity2.UserName)
	checkField(t, entity.fieldByName("`loginTime`"), "`loginTime`", "=", entity2.LoginTime, &entity2.LoginTime)
	t.Log("values: ", entity.Values())
	t.Log("fields: ", entity.ScanFields())
	t.Log("args: ", entity.ScanArgs())

	entity21 := &tabEntity21{UserID2: 21}
	entity21.UserID = 2
	err = entity.Parse(entity21)
	if err != nil {
		t.Error(err)
	}

	entity22 := tabEntity22{}
	err = entity.Parse(entity22)
	if err == nil {
		t.Error("empty table name should be error")
	}

}

func checkField(t *testing.T, entityField *field, name, filter string, value, address interface{}) {
	if entityField.name != name {
		t.Error("field name error: expect=", name, ", actual=", entityField.name)
	}
	if entityField.filter != filter {
		t.Error("field filter error: expect=", filter, ", actual=", entityField.filter)
	}
	if entityField.value != value {
		t.Error("field value error: expect=", value, ", actual=", entityField.value)
	}
	if entityField.address != address {
		t.Error("field address error: expect=", address, ", actual=", entityField.address)
	}
}

type TabEntityBase struct {
	Field1 string
}

func (s TabEntityBase) TableName() string {
	return "tabTest"
}

type TabEntity1Base struct {
	Field2 string `sql:"field2"`
}
type TabEntity1 struct {
	TabEntityBase
	TabEntity1Base

	ID        string    `json:"id"`
	UserID    uint64    `sql:"userId" auto:"true"`
	UserName  string    `sql:"userName"`
	LoginTime time.Time `sql:"loginTime"`
}

type TabEntity2Base struct {
	Field22 string `sql:"field22"`
}
type TabEntity2 struct {
	TabEntityBase
	TabEntity2Base

	ID        string      `json:"id"`
	UserID    uint64      `sql:"userId" auto:"true" filter:"in"`
	UserName  interface{} `sql:"userName" filter:"like"`
	LoginTime time.Time   `sql:"loginTime"`
}

func (s TabEntity2) TableName() string {
	return "tabTest2"
}

type partsBase struct {
	UserName2 string `sql:"userName2"`
}

type tabEntity21 struct {
	TabEntity2
	partsBase

	UserID2 uint64 `sql:"userId2" auto:"true" primary:"true"`
}

func (s tabEntity21) TableName() string {
	return "tabTest21"
}

type tabEntity22 struct {
	TabEntity2
	partsBase
}

func (s tabEntity22) TableName() string {
	return ""
}
