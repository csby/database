package mysql

import (
	"database/sql"
	"fmt"
	"github.com/csby/database/sqldb"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type mysql struct {
	connection sqldb.SqlConnection
}

func NewDatabase(conn sqldb.SqlConnection) sqldb.SqlDatabase {
	return &mysql{connection: conn}
}

func (s *mysql) Open() (*sql.DB, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *mysql) Test() (string, error) {
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
	db.QueryRow("SELECT VERSION()").Scan(&dbVer)

	return dbVer, nil
}

func (s *mysql) Tables() ([]*sqldb.SqlTable, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sb := &strings.Builder{}
	sb.WriteString("select `table_name`, `table_comment` ")
	sb.WriteString("from `information_schema`.`tables` ")
	sb.WriteString(fmt.Sprintf("where `table_schema`='%s' ", s.connection.SchemaName()))
	sb.WriteString("and `table_type` = 'BASE TABLE'")

	query := sb.String()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*sqldb.SqlTable, 0)
	name := ""
	description := ""
	for rows.Next() {
		err = rows.Scan(&name, &description)
		if err != nil {
			return nil, err
		}

		table := &sqldb.SqlTable{
			Name:        name,
			Description: description,
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (s *mysql) Views() ([]*sqldb.SqlTable, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sb := &strings.Builder{}
	sb.WriteString("select `table_name`, `table_comment` ")
	sb.WriteString("from `information_schema`.`tables` ")
	sb.WriteString(fmt.Sprintf("where `table_schema`='%s' ", s.connection.SchemaName()))
	sb.WriteString("and `table_type` = 'VIEW'")

	query := sb.String()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*sqldb.SqlTable, 0)
	name := ""
	description := ""
	for rows.Next() {
		err = rows.Scan(&name, &description)
		if err != nil {
			return nil, err
		}

		table := &sqldb.SqlTable{
			Name:        name,
			Description: description,
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (s *mysql) Columns(tableName string) ([]*sqldb.SqlColumn, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sb := &strings.Builder{}
	sb.WriteString("SELECT ")
	sb.WriteString("`column_name`, ")
	sb.WriteString("`column_type`, ")
	sb.WriteString("`column_comment`, ")
	sb.WriteString("`column_key`, ")
	sb.WriteString("`is_nullable`, ")
	sb.WriteString("`data_type`, ")
	sb.WriteString("`column_default`, ")
	sb.WriteString("`extra` ")

	sb.WriteString("from `information_schema`.`columns` ")
	sb.WriteString("where `table_schema`=? and `table_name`=? ")

	rows, err := db.Query(sb.String(), s.connection.SchemaName(), tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]*sqldb.SqlColumn, 0)
	columnName := ""
	columnType := ""
	var columnComment *string = nil
	columnKey := ""
	isNullable := ""
	dataType := ""
	extra := ""
	for rows.Next() {
		var dataDefault *string = nil
		err = rows.Scan(&columnName, &columnType, &columnComment, &columnKey, &isNullable, &dataType, &dataDefault, &extra)
		if err != nil {
			return nil, err
		}

		column := &sqldb.SqlColumn{
			Name:        columnName,
			Type:        columnType,
			DataType:    dataType,
			DataDefault: dataDefault,
		}
		if columnComment != nil {
			column.Comment = *columnComment
		}
		if strings.ToLower(extra) == "auto_increment" {
			column.AutoIncrement = true
		}
		if strings.ToLower(columnKey) == "pri" {
			column.PrimaryKey = true
		} else if strings.ToLower(columnKey) == "uni" {
			column.UniqueKey = true
		}
		if strings.ToLower(isNullable) == "yes" {
			column.Nullable = true
		}
		if dataDefault != nil {
			column.DataDisplay = *dataDefault
		}

		columns = append(columns, column)
	}

	return columns, nil
}

func (s *mysql) TableDefinition(table *sqldb.SqlTable) (string, error) {
	if table == nil {
		return "", fmt.Errorf("table is nil")
	}

	columns, err := s.Columns(table.Name)
	if err != nil {
		return "", err
	}
	columnCount := len(columns)
	if columnCount < 1 {
		return "", fmt.Errorf("no columns")
	}

	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", table.Name))
	sb.WriteString(fmt.Sprintln())

	sb.WriteString(fmt.Sprintf("CREATE TABLE `%s` (", table.Name))
	sb.WriteString(fmt.Sprintln())

	primaryKeys := make([]string, 0)
	uniqueKeys := make([]string, 0)
	for i := 0; i < columnCount; i++ {
		column := columns[i]
		sb.WriteString(fmt.Sprintf("`%s` %s ", column.Name, column.Type))
		if !column.Nullable {
			sb.WriteString("NOT NULL ")
		}
		if column.AutoIncrement {
			sb.WriteString("AUTO_INCREMENT ")
		}
		if column.DataDefault != nil {
			sb.WriteString(fmt.Sprintf("DEFAULT '%s' ", *column.DataDefault))
		}
		if len(column.Comment) > 0 {
			sb.WriteString(fmt.Sprintf("COMMENT '%s' ", column.Comment))
		}
		if i < columnCount-1 {
			sb.WriteString(",")
			sb.WriteString(fmt.Sprintln())
		}

		if column.PrimaryKey {
			primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", column.Name))
		}
		if column.UniqueKey {
			uniqueKeys = append(uniqueKeys, fmt.Sprintf("%s", column.Name))
		}
	}

	if len(primaryKeys) > 0 {
		sb.WriteString(", ")
		sb.WriteString(fmt.Sprintln())
		sb.WriteString(fmt.Sprintf("PRIMARY KEY (%s) ", strings.Join(primaryKeys, ",")))
	}

	uniqueKeyCount := len(uniqueKeys)
	for i := 0; i < uniqueKeyCount; i++ {
		sb.WriteString(",")
		sb.WriteString(fmt.Sprintln())
		sb.WriteString(fmt.Sprintf("UNIQUE KEY `%s_UNIQUE` (`%s`) ", uniqueKeys[i], uniqueKeys[i]))
	}

	sb.WriteString(fmt.Sprintln())
	sb.WriteString(") ")
	if len(table.Description) > 0 {
		sb.WriteString(fmt.Sprintf("COMMENT='%s'", table.Description))
	}
	sb.WriteString(fmt.Sprintln())

	return sb.String(), nil
}

func (s *mysql) ViewDefinition(viewName string) (string, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return "", err
	}
	defer db.Close()

	tableSchema := s.connection.SchemaName()
	sb := &strings.Builder{}
	sb.WriteString("select `view_definition` ")
	sb.WriteString("from `information_schema`.`views` ")
	sb.WriteString(fmt.Sprintf("where `table_schema`='%s' ", tableSchema))
	sb.WriteString(fmt.Sprintf("and `table_name`='%s' ", viewName))

	query := sb.String()
	row := db.QueryRow(query)

	definition := ""
	err = row.Scan(&definition)
	if err != nil {
		return "", err
	}

	definition = strings.Replace(definition, fmt.Sprintf("`%s`.", tableSchema), "", -1)

	return fmt.Sprintf("CREATE OR REPLACE VIEW `%s` As %s", viewName, definition), nil
}

func (s *mysql) NewAccess(transactional bool) (sqldb.SqlAccess, error) {
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

func (s *mysql) NewEntity() sqldb.SqlEntity {
	return &entity{}
}

func (s *mysql) NewBuilder() sqldb.SqlBuilder {
	instance := &builder{}
	instance.Reset()

	return instance
}

func (s *mysql) NewFilter(entity interface{}, fieldOr, groupOr bool) sqldb.SqlFilter {
	return newFilter(entity, fieldOr, groupOr)
}

func (s *mysql) IsNoRows(err error) bool {
	if err == nil {
		return false
	}

	if err == sql.ErrNoRows {
		return true
	}

	return false
}

func (s *mysql) Insert(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.Insert(entity)
}

func (s *mysql) InsertSelective(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.InsertSelective(entity)
}

func (s *mysql) Delete(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.Delete(entity, filters...)
}

func (s *mysql) Update(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.Update(entity, filters...)
}

func (s *mysql) UpdateSelective(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.UpdateSelective(entity, filters...)
}

func (s *mysql) UpdateByPrimaryKey(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.UpdateByPrimaryKey(entity)
}

func (s *mysql) UpdateSelectiveByPrimaryKey(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.UpdateSelectiveByPrimaryKey(entity)
}

func (s *mysql) SelectOne(entity interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectOne(entity, filters...)
}

func (s *mysql) SelectDistinct(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectDistinct(entity, row, order, filters...)
}

func (s *mysql) SelectList(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectList(entity, row, order, filters...)
}

func (s *mysql) SelectPage(entity interface{}, page func(total, page, size, index uint64), row func(index uint64, evt sqldb.SqlEvent), size, index uint64, order interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectPage(entity, page, row, size, index, order, filters...)
}

func (s *mysql) SelectCount(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectCount(entity, filters...)
}
