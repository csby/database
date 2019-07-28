package mssql

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
	db := &mssql{
		connection: testConnection(),
	}
	t.Log("connection:", db.connection.SourceName())

	dbVer, err := db.Test()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("version: ", dbVer)
}

func TestMssql_Tables(t *testing.T) {
	db := &mssql{
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

func TestMssql_Views(t *testing.T) {
	db := &mssql{
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

func TestMssql_Columns(t *testing.T) {
	db := &mssql{
		connection: testConnection(),
	}
	tableName := "Admission"
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

func TestMssql_TableDefinition(t *testing.T) {
	db := &mssql{
		connection: testConnection(),
	}
	table := &sqldb.SqlTable{
		Name:        "AdmissionTest",
		Description: "dd",
	}
	definition, err := db.TableDefinition(table)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("definition:", definition)
}

func TestMssql_ViewDefinition(t *testing.T) {
	db := &mssql{
		connection: testConnection(),
	}
	viewName := "ViewAdmission"
	definition, err := db.ViewDefinition(viewName)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("definition:", definition)
}

func TestMssql_Version(t *testing.T) {
	db := &mssql{
		connection: testConnection(),
	}

	t.Log("version: ", db.Version())
}

func TestMssql_SelectList(t *testing.T) {
	db := &mssql{
		connection: testConnection(),
	}
	dbEntity := &tabEntityUser{}
	err := db.SelectList(dbEntity, func(index uint64, evt sqldb.SqlEvent) {
		t.Log(fmt.Sprintf("%3d ", index), "UserId:", dbEntity.UserId, "; UserName:", dbEntity.UserName)
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMssql_SelectPage(t *testing.T) {
	db := &mssql{
		connection: testConnection(),
	}
	dbEntity := &tabEntityUser{}
	dbOrder := &tabEntityUserOrder{}
	dbFilter := &tabEntityUserFilter{
		Auth: 0,
	}
	sqlFilter := db.NewFilter(dbFilter, false, false)
	err := db.SelectPage(dbEntity, func(total, page, size, index uint64) {
		t.Log("total:", total, "; page:", page, "; size:", size, "; index:", index)
	}, func(index uint64, evt sqldb.SqlEvent) {
		t.Log("UserId:", dbEntity.UserId, "; UserName:", dbEntity.UserName)
	}, 3, 2, dbOrder, sqlFilter)
	if err != nil {
		t.Fatal(err)
	}
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
	cfgPath := filepath.Join(goPath, "tmp", "cfg", "database_mssql_test.json")
	cfg := &Connection{
		Host:     "127.0.0.1",
		Port:     1433,
		Schema:   "test",
		Instance: "MSSQLSERVER",
		User:     "sa",
		Password: "",
		Timeout:  10,
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

type TabEntityBase struct {
}

func (s TabEntityBase) TableName() string {
	return "User"
}

type tabEntityUser struct {
	TabEntityBase

	UserId   string `sql:"UserId" primary:"true"`
	UserName string `sql:"UserName"`
}

type tabEntityUserOrder struct {
	TabEntityBase

	UserName string `sql:"UserName" order:"DESC"`
}

type tabEntityUserFilter struct {
	TabEntityBase

	Auth uint64 `sql:"Auth"`
}
