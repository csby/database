package mysql

import (
	"fmt"
	"github.com/csby/database/sqldb"
	"os"
	"path/filepath"
	"runtime"
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
	table := &sqldb.SqlTable{
		Name: "DoctorUserAuths",
	}
	columns, err := db.Columns(table)
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
	paths := strings.Split(goPath, string(os.PathListSeparator))
	if len(paths) > 1 {
		goPath = paths[0]

		_, file, _, _ := runtime.Caller(0)
		fileDir := strings.ToLower(filepath.Dir(file))
		for _, path := range paths {
			if strings.HasPrefix(fileDir, strings.ToLower(path)) {
				goPath = path
				break
			}
		}
	}
	cfgPath := filepath.Join(goPath, "tmp", "cfg", "database_mysql_test.json")
	cfg := &Connection{
		Host:     "172.0.0.1",
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
