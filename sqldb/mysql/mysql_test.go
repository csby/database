package mysql

import (
	"fmt"
	"github.com/csby/database/sqldb"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTest(t *testing.T) {
	db := NewDatabase(testConnection())

	dbVer, err := db.Test()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("version: ", dbVer)
}

func TestMysql_Tables(t *testing.T) {
	db := &mysql{
		connection: testConnection(),
	}
	tables, err := db.Tables()
	if err != nil {
		t.Fatal(err)
	}
	count := len(tables)
	t.Log("count:", count)
	for i := 0; i < count; i++ {
		t.Logf("%2d %+v", i+1, tables[i])
	}
}

func TestMysql_Views(t *testing.T) {
	db := &mysql{
		connection: testConnection(),
	}
	views, err := db.Views()
	if err != nil {
		t.Fatal(err)
	}
	count := len(views)
	t.Log("count:", count)
	for i := 0; i < count; i++ {
		t.Logf("%2d %+v", i+1, views[i])
	}
}

func TestMysql_Columns(t *testing.T) {
	db := &mysql{
		connection: testConnection(),
	}
	tableName := "DoctorUserAuths"
	columns, err := db.Columns(tableName)
	if err != nil {
		t.Fatal(err)
	}
	count := len(columns)
	t.Log("count:", count)
	for i := 0; i < count; i++ {
		t.Logf("%2d %+v", i+1, columns[i])
	}
}

func TestMysql_TableDefinition(t *testing.T) {
	db := &mysql{
		connection: testConnection(),
	}
	table := &sqldb.SqlTable{
		Name:        "AlertRecord",
		Description: "dd",
	}
	definition, err := db.TableDefinition(table)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("definition:", definition)
}

func TestMysql_ViewDefinition(t *testing.T) {
	db := &mysql{
		connection: testConnection(),
	}
	viewName := "ViewAlertRecord"
	definition, err := db.ViewDefinition(viewName)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("definition:", definition)
}

func testConnection() *Connection {
	goPath := os.Getenv("GOPATH")
	paths := strings.Split(goPath, ";")
	if len(paths) < 2 {
		paths = strings.Split(goPath, ":")
	}
	if len(paths) > 1 {
		goPath = paths[0]
	}
	cfgPath := filepath.Join(goPath, "tmp", "cfg", "database_mysql_test.json")
	cfg := &Connection{
		Server:   "172.0.0.1",
		Port:     3306,
		Schema:   "mysql",
		Charset:  "utf8",
		Timeout:  10,
		User:     "root",
		Password: "",
	}
	_, err := os.Stat(cfgPath)
	if os.IsNotExist(err) {
		err = cfg.SaveToFile(cfgPath)
		if err != nil {
			fmt.Println("generate configure file fail: ", err)
		}
	} else {
		err = cfg.LoadFromFile(cfgPath)
		if err != nil {
			fmt.Println("load configure file fail: ", err)
		}
	}

	return cfg
}
