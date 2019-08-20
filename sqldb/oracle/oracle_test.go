package oracle

import (
	"fmt"
	"github.com/csby/database/sqldb"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOracle_Test(t *testing.T) {
	db := &Oracle{
		connection: testConnection(),
	}
	t.Log("connection:", db.connection.SourceName())

	dbVer, err := db.Test()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("version: ", dbVer)
}

func TestOracle_Tables(t *testing.T) {
	db := &Oracle{
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

func TestOracle_Views(t *testing.T) {
	db := &Oracle{
		connection: testConnection(),
	}
	tables, err := db.Views()
	if err != nil {
		t.Fatal(err)
	}
	count := len(tables)
	t.Log("count:", count)
	for i := 0; i < count; i++ {
		t.Logf("%2d %+v", i+1, tables[i])
	}
}

func TestOracle_Columns(t *testing.T) {
	db := &Oracle{
		connection: testConnection(),
	}
	tableName := "LAB.ANTIBIOTICS_RESULT_REFER"
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

func TestOracle_getOwnerAndName(t *testing.T) {
	db := &Oracle{}
	owner, name := db.getOwnerAndName("EXAM.EXAM_IMAGE_INDEX")
	t.Log("owner:", owner)
	t.Log("name :", name)
	owner, name = db.getOwnerAndName("EXAM.")
	t.Log("owner:", owner)
	t.Log("name :", name)
	owner, name = db.getOwnerAndName("EXAM")
	t.Log("owner:", owner)
	t.Log("name :", name)
}

func TestOracle_SelectList(t *testing.T) {
	db := &Oracle{
		connection: testConnection(),
	}

	dbFilter := &TabEntityFilter{
		AntibioticsCode: "4",
	}
	sqlFilter := db.NewFilter(dbFilter, false, false)

	dbEntity := &TabEntity{}
	err := db.SelectList(dbEntity, func(index uint64, evt sqldb.SqlEvent) {
		t.Log(fmt.Sprintf("%3d ", index), "AntibioticsCode:", dbEntity.AntibioticsCode, "; TestMethod:", dbEntity.TestMethod)
	}, nil, sqlFilter)
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
	cfgPath := filepath.Join(goPath, "tmp", "cfg", "database_oracle_test.json")
	cfg := &Connection{
		Host:     "127.0.0.1",
		Port:     1521,
		SID:      "orcl",
		User:     "dba",
		Password: "***",
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
	return "LAB.ANTIBIOTICS_RESULT_REFER"
}

type TabEntity struct {
	TabEntityBase
	//
	AntibioticsCode string `sql:"ANTIBIOTICS_CODE"`
	//
	TestMethod string `sql:"TEST_METHOD"`
}

type TabEntityFilter struct {
	TabEntityBase
	//
	AntibioticsCode string `sql:"ANTIBIOTICS_CODE"`
}
