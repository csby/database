package sqldb

import "database/sql"

type SqlFactory interface {
	NewDatabase() SqlDatabase
}

type SqlConnection interface {
	DriverName() string
	SourceName() string
	SchemaName() string
	ClusterSourceName(readOnly bool) string
}

type SqlDatabase interface {
	Instances(host, port string) ([]SqlInstance, error)
	Test() (string, error)
	ClusterTest(readOnly bool) (string, error)
	Schema() string
	Tables() ([]*SqlTable, error)
	Views() ([]*SqlTable, error)
	Columns(table *SqlTable) ([]*SqlColumn, error)

	NewAccess(transactional bool) (SqlAccess, error)
	NewClusterAccess(transactional bool, readOnly bool) (SqlAccess, error)
	NewEntity() SqlEntity
	NewBuilder() SqlBuilder
	NewFilter(entity interface{}, fieldOr, groupOr bool) SqlFilter

	IsNoRows(err error) bool
	Insert(entity interface{}) (uint64, error)
	InsertSelective(entity interface{}) (uint64, error)
	Delete(entity interface{}, filters ...SqlFilter) (uint64, error)
	Update(entity interface{}, filters ...SqlFilter) (uint64, error)
	UpdateSelective(entity interface{}, filters ...SqlFilter) (uint64, error)
	UpdateByPrimaryKey(entity interface{}) (uint64, error)
	UpdateSelectiveByPrimaryKey(entity interface{}) (uint64, error)
	SelectCount(entity interface{}, filters ...SqlFilter) (uint64, error)
	SelectOne(entity interface{}, filters ...SqlFilter) error
	SelectDistinct(entity interface{}, row func(idx uint64, evt SqlEvent), order interface{}, filters ...SqlFilter) error
	SelectList(entity interface{}, row func(idx uint64, evt SqlEvent), order interface{}, filters ...SqlFilter) error
	SelectPage(entity interface{}, page func(total, page, size, index uint64), row func(idx uint64, evt SqlEvent), size, index uint64, order interface{}, filters ...SqlFilter) error
}

type SqlInstance interface {
	Name() string
	Port() string
	Version() string
}

type SqlAccess interface {
	Close() error
	Commit() error
	Version() int

	NewFilter(entity interface{}, fieldOr, groupOr bool) SqlFilter

	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row

	IsNoRows(err error) bool
	Insert(entity interface{}, fields ...SqlField) (uint64, error)
	InsertSelective(entity interface{}) (uint64, error)
	Delete(entity interface{}, filters ...SqlFilter) (uint64, error)
	Update(entity interface{}, filters ...SqlFilter) (uint64, error)
	UpdateSelective(entity interface{}, filters ...SqlFilter) (uint64, error)
	UpdateByPrimaryKey(entity interface{}) (uint64, error)
	UpdateSelectiveByPrimaryKey(entity interface{}) (uint64, error)
	SelectCount(entity interface{}, filters ...SqlFilter) (uint64, error)
	SelectOne(entity interface{}, filters ...SqlFilter) error
	SelectDistinct(entity interface{}, row func(idx uint64, evt SqlEvent), order interface{}, filters ...SqlFilter) error
	SelectList(entity interface{}, row func(idx uint64, evt SqlEvent), order interface{}, filters ...SqlFilter) error
	SelectPage(entity interface{}, page func(total, page, size, index uint64), row func(idx uint64, evt SqlEvent), size, index uint64, order interface{}, filters ...SqlFilter) error
}

type SqlEvent interface {
	Cancel(err error)
}

type SqlField interface {
	Name() string
	Value() interface{}
	Address() interface{}
	AutoIncrement() bool
	PrimaryKey() bool
	Filter() string
	Order() string
	ValueEmpty() bool
}

type SqlEntity interface {
	Parse(entity interface{}) error
	ParseFilter(entity interface{}) error
	Name() string
	FieldCount() int
	Field(i int) SqlField
	ScanFields() string
	ScanArgs() []interface{}
	Values() []interface{}
}

type SqlBuilder interface {
	Query() string
	Args() []interface{}
	ArgName() string

	Reset() SqlBuilder
	Select(query string, distinct bool) SqlBuilder
	Insert(query string) SqlBuilder
	Delete(query string) SqlBuilder
	Update(query string) SqlBuilder
	From(query string) SqlBuilder
	Value(filed string, value interface{}) SqlBuilder
	Set(filed string, value interface{}) SqlBuilder
	WhereFormatAnd(format string, a ...interface{}) SqlBuilder
	WhereFormatOr(format string, a ...interface{}) SqlBuilder
	WhereFormat(format string, a ...interface{}) SqlBuilder
	WhereAnd(query string, args ...interface{}) SqlBuilder
	WhereOr(query string, args ...interface{}) SqlBuilder
	Where(query string, args ...interface{}) SqlBuilder
	Order(query string) SqlBuilder
	Append(query string, args ...interface{}) SqlBuilder
	AppendFormat(format string, a ...interface{}) SqlBuilder
}

type SqlFilter interface {
	FieldOr() bool
	GroupOr() bool
	Fields() interface{}
}

type SqlName struct {
	Name string `json:"name"`
}

type SqlTable struct {
	Schema      string `json:"schema"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Rows        int64  `json:"rows"`
}

type SqlColumn struct {
	Id      int    `json:"id" note:"ID"`
	Name    string `json:"name" note:"名称"`
	Type    string `json:"type" note:"类型"`
	Comment string `json:"comment" note:"说明"`

	Nullable      bool `json:"nullable" note:"是否可空"`
	PrimaryKey    bool `json:"primaryKey" note:"是否主键"`
	UniqueKey     bool `json:"uniqueKey" note:"是否唯一"`
	AutoIncrement bool `json:"autoIncrement" note:"是否自增长"`

	DataType    string  `json:"dataType" note:"数据类型"`
	Precision   *int    `json:"precision" note:"精度"`
	Scale       *int    `json:"scale" note:"小数点"`
	DataDefault *string `json:"dataDefault" note:"数据默认值"`
	DataDisplay string  `json:"dataDisplay" note:"数据默认值显示"`
}

type SqlIndex struct {
	Id         int          `json:"id" note:"ID"`
	Name       string       `json:"name" note:"名称"`
	Type       int          `json:"type" note:"类型: 1-聚集; 2-非聚集"`
	PrimaryKey bool         `json:"primaryKey" note:"是否主键"`
	UniqueKey  bool         `json:"uniqueKey" note:"是否唯一"`
	Fragment   float32      `json:"fragment" note:"碎片总计"`
	Columns    []*SqlColumn `json:"columns" note:"包含列"`
}

type SqlTableCount struct {
	Name string `json:"name" note:"表名称"`
	Rows int64  `json:"rows" note:"记录行数"`
}
