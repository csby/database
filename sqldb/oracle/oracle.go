package oracle

import (
	"database/sql"
	"fmt"
	"github.com/csby/database/sqldb"
	"strings"

	_ "gopkg.in/goracle.v2"
)

type Oracle struct {
	connection sqldb.SqlConnection
}

func NewDatabase(conn sqldb.SqlConnection) sqldb.SqlDatabase {
	return &Oracle{connection: conn}
}

func (s *Oracle) Open() (*sql.DB, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *Oracle) Test() (string, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return "", err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return "", err
	}

	dbVer := ""
	db.QueryRow("SELECT banner FROM v$version").Scan(&dbVer)
	index := strings.Index(dbVer, "\n")
	if index > 0 {
		dbVer = dbVer[0:index]
	}

	return dbVer, nil
}

func (s *Oracle) Tables() ([]*sqldb.SqlTable, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sb := &strings.Builder{}
	sb.WriteString("select t.owner, t.table_name, t.comments ")
	sb.WriteString("from all_tab_comments t ")
	sb.WriteString("where t.table_type='TABLE' ")

	conn, ok := s.connection.(*Connection)
	if ok {
		if len(conn.Owners) > 0 {
			owners := strings.Join(conn.Owners, "','")
			sb.WriteString("and t.owner in ('")
			sb.WriteString(owners)
			sb.WriteString("') ")
		}
	}

	query := sb.String()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*sqldb.SqlTable, 0)
	owner := ""
	name := ""
	var description *string = nil
	for rows.Next() {
		err = rows.Scan(&owner, &name, &description)
		if err != nil {
			return nil, err
		}

		table := &sqldb.SqlTable{
			Name: fmt.Sprintf("%s.%s", owner, name),
		}
		if description != nil {
			table.Description = *description
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (s *Oracle) Views() ([]*sqldb.SqlTable, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sb := &strings.Builder{}
	sb.WriteString("select t.owner, t.table_name, t.comments ")
	sb.WriteString("from all_tab_comments t ")
	sb.WriteString("where t.table_type='VIEW' ")

	conn, ok := s.connection.(*Connection)
	if ok {
		if len(conn.Owners) > 0 {
			owners := strings.Join(conn.Owners, "','")
			sb.WriteString("and t.owner in ('")
			sb.WriteString(owners)
			sb.WriteString("') ")
		}
	}

	query := sb.String()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*sqldb.SqlTable, 0)
	owner := ""
	name := ""
	var description *string = nil
	for rows.Next() {
		err = rows.Scan(&owner, &name, &description)
		if err != nil {
			return nil, err
		}

		table := &sqldb.SqlTable{
			Name: fmt.Sprintf("%s.%s", owner, name),
		}
		if description != nil {
			table.Description = *description
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (s *Oracle) Columns(tableName string) ([]*sqldb.SqlColumn, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tabOwner, tabName := s.getOwnerAndName(tableName)

	// 列名 | 列说明 | 数据类型 | 长度 | 精度 | 小数位数 | 允许空
	sb := &strings.Builder{}
	sb.WriteString("select ")
	sb.WriteString("t1.column_name, ")
	sb.WriteString("t1.data_type, ")
	sb.WriteString("t1.data_length, ")
	sb.WriteString("t1.data_precision, ")
	sb.WriteString("t1.data_scale, ")
	sb.WriteString("t1.nullable, ")
	sb.WriteString("t2.comments ")

	sb.WriteString("from all_tab_cols t1 ")
	sb.WriteString("left join all_col_comments t2 on t1.owner = t2.owner ")
	sb.WriteString("and t1.table_name = t2.table_name ")
	sb.WriteString("and t1.column_name = t2.column_name ")
	sb.WriteString("where t1.table_name = '")
	sb.WriteString(tabName)
	sb.WriteString("' ")
	if len(tabOwner) > 0 {
		sb.WriteString("and t1.owner = '")
		sb.WriteString(tabOwner)
		sb.WriteString("' ")
	}

	query := sb.String()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]*sqldb.SqlColumn, 0)
	name := ""
	dataType := ""
	length := 0
	nullable := ""
	for rows.Next() {
		var comment *string = nil
		var precision *int = nil
		var scale *int = nil
		err = rows.Scan(&name, &dataType, &length, &precision, &scale, &nullable, &comment)
		if err != nil {
			return nil, err
		}

		column := &sqldb.SqlColumn{
			Name:      name,
			DataType:  dataType,
			Precision: precision,
			Scale:     scale,
		}
		if comment != nil {
			column.Comment = *comment
		}
		if nullable == "Y" {
			column.Nullable = true
		}
		column.Type = s.columnTypeName(dataType, length, precision, scale)

		columns = append(columns, column)
	}

	return columns, nil
}

func (s *Oracle) TableDefinition(table *sqldb.SqlTable) (string, error) {
	return "", fmt.Errorf("not implement")
}

func (s *Oracle) ViewDefinition(viewName string) (string, error) {
	return "", fmt.Errorf("not implement")
}

func (s *Oracle) NewAccess(transactional bool) (sqldb.SqlAccess, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}

	if transactional {
		tx, err := db.Begin()
		if err != nil {
			db.Close()
			return nil, err
		}

		return &transaction{db: db, tx: tx}, nil
	}

	return &normal{db: db}, nil
}

func (s *Oracle) NewEntity() sqldb.SqlEntity {
	return &entity{}
}

func (s *Oracle) NewBuilder() sqldb.SqlBuilder {
	instance := &builder{}
	instance.Reset()

	return instance
}

func (s *Oracle) NewFilter(entity interface{}, fieldOr, groupOr bool) sqldb.SqlFilter {
	return newFilter(entity, fieldOr, groupOr)
}

func (s *Oracle) IsNoRows(err error) bool {
	if err == nil {
		return false
	}

	if err == sql.ErrNoRows {
		return true
	}

	return false
}

func (s *Oracle) Insert(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.Insert(entity)
}

func (s *Oracle) InsertSelective(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.InsertSelective(entity)
}

func (s *Oracle) Delete(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.Delete(entity, filters...)
}

func (s *Oracle) Update(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.Update(entity, filters...)
}

func (s *Oracle) UpdateSelective(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.UpdateSelective(entity, filters...)
}

func (s *Oracle) UpdateByPrimaryKey(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.UpdateByPrimaryKey(entity)
}

func (s *Oracle) UpdateSelectiveByPrimaryKey(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.UpdateSelectiveByPrimaryKey(entity)
}

func (s *Oracle) SelectOne(entity interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectOne(entity, filters...)
}

func (s *Oracle) SelectDistinct(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectDistinct(entity, row, order, filters...)
}

func (s *Oracle) SelectList(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectList(entity, row, order, filters...)
}

func (s *Oracle) SelectPage(entity interface{}, page func(total, page, size, index uint64), row func(index uint64, evt sqldb.SqlEvent), size, index uint64, order interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectPage(entity, page, row, size, index, order, filters...)
}

func (s *Oracle) SelectCount(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectCount(entity, filters...)
}

func (s *Oracle) getOwnerAndName(name string) (string, string) {
	index := strings.Index(name, ".")
	if index > 0 {
		return string(name[0:index]), string(name[index+1:])
	} else {
		return "", name
	}
}

func (s *Oracle) columnTypeName(dataType string, length int, precision, scale *int) string {
	sb := &strings.Builder{}
	sb.WriteString(dataType)
	if scale != nil && precision != nil {
		sb.WriteString(fmt.Sprintf("(%d, %d)", *precision, *scale))
	} else {
		sb.WriteString(fmt.Sprintf("(%d)", length))
	}

	return sb.String()
}
