package mssql

import (
	"database/sql"
	"fmt"
	"github.com/csby/database/sqldb"
	"strconv"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
)

type mssql struct {
	connection sqldb.SqlConnection
}

func NewDatabase(conn sqldb.SqlConnection) sqldb.SqlDatabase {
	return &mssql{connection: conn}
}

func (s *mssql) Open() (*sql.DB, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *mssql) Instances(host, port string) ([]sqldb.SqlInstance, error) {
	instances, err := getInstances(host, port)
	if err != nil {
		return nil, err
	}

	results := make([]sqldb.SqlInstance, 0)
	for name, value := range instances {
		if len(name) < 1 {
			continue
		}

		val, ok := value["tcp"]
		if !ok {
			continue
		}
		result := &Instance{
			name: name,
			port: val,
		}

		val, ok = value["Version"]
		if ok {
			result.version = val
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *mssql) Test() (string, error) {
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
	db.QueryRow("SELECT @@VERSION").Scan(&dbVer)
	index := strings.Index(dbVer, "\n")
	if index > 0 {
		dbVer = dbVer[0:index]
	}

	return dbVer, nil
}

func (s *mssql) Version() int {
	version, err := s.Test()
	if err != nil {
		return 0
	}

	vs := strings.Split(version, " ")
	if len(vs) > 3 {
		v, err := strconv.Atoi(vs[3])
		if err == nil {
			return v
		}
	}

	return 0
}

func (s *mssql) Tables() ([]*sqldb.SqlTable, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sb := &strings.Builder{}
	sb.WriteString("select s.[name], t.[name], e.[value] ")
	sb.WriteString("from [sys].[tables] t ")
	sb.WriteString("left join [sys].[schemas] s on s.[schema_id] = t.[schema_id]")
	sb.WriteString("left join [sys].[extended_properties] e on e.[major_id] = t.[object_id] and e.[minor_id] = 0 ")

	query := sb.String()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*sqldb.SqlTable, 0)
	schema := ""
	name := ""
	var description *string = nil
	for rows.Next() {
		err = rows.Scan(&schema, &name, &description)
		if err != nil {
			return nil, err
		}

		table := &sqldb.SqlTable{
			Schema: schema,
			Name:   name,
		}
		if description != nil {
			table.Description = *description
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (s *mssql) Views() ([]*sqldb.SqlTable, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sb := &strings.Builder{}
	sb.WriteString("select s.[name], v.[name] ")
	sb.WriteString("from [sys].[views] v ")
	sb.WriteString("left join [sys].[schemas] s on s.[schema_id] = v.[schema_id] ")

	query := sb.String()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*sqldb.SqlTable, 0)
	schema := ""
	name := ""
	for rows.Next() {
		err = rows.Scan(&schema, &name)
		if err != nil {
			return nil, err
		}

		table := &sqldb.SqlTable{
			Schema: schema,
			Name:   name,
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (s *mssql) Columns(table *sqldb.SqlTable) ([]*sqldb.SqlColumn, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 列名 | 列说明 | 数据类型 | 长度 | 精度 | 小数位数 | 标识 | 主键 | 允许空 | 默认值
	sqlStr := `
SELECT  
        col.[name] AS 列名 ,  
        ISNULL(ep.[value], '') AS 列说明 ,  
        t.[name] AS 数据类型 ,  
        col.[length] AS 长度 ,  
		col.[xprec] AS 精度,
        COLUMNPROPERTY(col.[id], col.[name], 'Scale') AS 小数位数 , 
		COLUMNPROPERTY(col.[id], col.[name], 'IsIdentity') AS 标识 ,  
        CASE WHEN EXISTS ( SELECT   1  
                           FROM     [dbo].[sysindexes] si  
                                    INNER JOIN [dbo].[sysindexkeys] sik ON si.[id] = sik.[id]  
                                                              AND si.[indid] = sik.[indid]  
                                    INNER JOIN [dbo].[syscolumns] sc ON sc.[id] = sik.[id]  
                                                              AND sc.[colid] = sik.[colid]  
                                    INNER JOIN [dbo].[sysobjects] so ON so.[name] = si.[name]  
                                                              AND so.[xtype] = 'PK'  
                           WHERE    sc.[id] = col.[id]  
                                    AND sc.[colid] = col.[colid] ) THEN 1  
             ELSE 0  
        END AS 主键 ,  
		col.[isnullable] AS 允许空 ,  
        comm.[text] AS 默认值  
FROM    [dbo].[syscolumns] col  
        LEFT  JOIN [dbo].[systypes] t ON col.[xtype] = t.[xusertype]  
        inner JOIN [dbo].[sysobjects] obj ON col.[id] = obj.[id]  
                                         AND (obj.[xtype] = 'U'  or obj.[xtype] = 'V') 
                                         AND obj.[status] >= 0  
        LEFT  JOIN [dbo].[syscomments] comm ON col.[cdefault] = comm.id  
        LEFT  JOIN [sys].[extended_properties] ep ON col.[id] = ep.[major_id]  
                                                      AND col.[colid] = ep.[minor_id]  
                                                      AND ep.[name] = 'MS_Description'  
        LEFT  JOIN [sys].[extended_properties] epTwo ON obj.[id] = epTwo.[major_id]  
                                                         AND epTwo.[minor_id] = 0  
                                                         AND epTwo.[name] = 'MS_Description'  
		LEFT  JOIN [sys].[tables] tab ON tab.[object_id] = obj.[id] 
		LEFT  JOIN [sys].[views] vie ON vie.[object_id] = obj.[id] 
		LEFT  JOIN [sys].[schemas] sch ON (sch.[schema_id] = tab.[schema_id] or sch.[schema_id] = vie.[schema_id])
WHERE   obj.[name] = 
`

	sb := &strings.Builder{}
	sb.WriteString(sqlStr)
	sb.WriteString("'")
	sb.WriteString(table.Name)
	sb.WriteString("'")

	sb.WriteString(" and  sch.[name] = '")
	sb.WriteString(table.Schema)
	sb.WriteString("'")

	query := sb.String()
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]*sqldb.SqlColumn, 0)
	name := ""
	var comment *string = nil
	dataType := ""
	length := 0
	autoIncrement := 0
	primaryKey := 0
	nullable := 0
	for rows.Next() {
		var precision *int = nil
		var scale *int = nil
		var dataDefault *string = nil
		err = rows.Scan(&name, &comment, &dataType, &length, &precision, &scale, &autoIncrement, &primaryKey, &nullable, &dataDefault)
		if err != nil {
			return nil, err
		}

		column := &sqldb.SqlColumn{
			Name:        name,
			DataType:    dataType,
			DataDefault: dataDefault,
			Precision:   precision,
			Scale:       scale,
		}
		if comment != nil {
			column.Comment = *comment
		}
		if autoIncrement > 0 {
			column.AutoIncrement = true
		}
		if primaryKey > 0 {
			column.PrimaryKey = true
		}
		if nullable > 0 {
			column.Nullable = true
		}
		if dataDefault != nil {
			column.DataDisplay = s.columnDateDefault(*dataDefault)
		}
		column.Type = s.columnTypeName(dataType, length, precision, scale)

		columns = append(columns, column)
	}

	return columns, nil
}

func (s *mssql) TableDefinition(table *sqldb.SqlTable) (string, error) {
	if table == nil {
		return "", fmt.Errorf("table is nil")
	}

	columns, err := s.Columns(table)
	if err != nil {
		return "", err
	}
	columnCount := len(columns)
	if columnCount < 1 {
		return "", fmt.Errorf("no columns")
	}

	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("IF OBJECT_ID('[dbo].[%s]') IS NOT NULL", table.Name))
	sb.WriteString(fmt.Sprintln())
	sb.WriteString(fmt.Sprintf("	DROP TABLE [dbo].[%s]", table.Name))
	sb.WriteString(fmt.Sprintln())
	sb.WriteString(fmt.Sprintln("GO"))

	sb.WriteString(fmt.Sprintln())
	sb.WriteString(fmt.Sprintf("CREATE TABLE [dbo].[%s] (", table.Name))
	sb.WriteString(fmt.Sprintln())

	primaryKeys := make([]string, 0)
	uniqueKeys := make([]string, 0)
	sbDefaults := &strings.Builder{}
	sbComments := &strings.Builder{}
	for i := 0; i < columnCount; i++ {
		column := columns[i]
		sb.WriteString(fmt.Sprintf("	[%s] %s ", column.Name, column.Type))
		if column.AutoIncrement {
			sb.WriteString("IDENTITY ")
		}
		if !column.Nullable {
			sb.WriteString("NOT NULL ")
		}
		if i < columnCount-1 {
			sb.WriteString(",")
			sb.WriteString(fmt.Sprintln())
		}

		if column.PrimaryKey {
			primaryKeys = append(primaryKeys, fmt.Sprintf("%s", column.Name))
		}
		if column.UniqueKey {
			uniqueKeys = append(uniqueKeys, fmt.Sprintf("%s", column.Name))
		}
		if column.DataDefault != nil {
			sbDefaults.WriteString(fmt.Sprintln())
			sbDefaults.WriteString(fmt.Sprintf("ALTER TABLE [dbo].[%[1]s] ADD  CONSTRAINT [DF_%[1]s_%[2]s]  DEFAULT %[3]s FOR [%[2]s] ",
				table.Name, column.Name, *column.DataDefault))
			sbDefaults.WriteString(fmt.Sprintln())
			sbDefaults.WriteString(fmt.Sprintln("GO"))
		}
		if len(column.Comment) > 0 {
			sbComments.WriteString(fmt.Sprintln(""))
			sbComments.WriteString(fmt.Sprintln("EXEC [sys].[sp_addextendedproperty] "))
			sbComments.WriteString(fmt.Sprintln("	@name=N'MS_Description', "))
			sbComments.WriteString(fmt.Sprintf("	@value=N'%s', ", column.Comment))
			sbComments.WriteString(fmt.Sprintln())
			sbComments.WriteString(fmt.Sprintln("	@level0type=N'SCHEMA', "))
			sbComments.WriteString(fmt.Sprintln("	@level0name=N'dbo', "))
			sbComments.WriteString(fmt.Sprintln("	@level1type=N'TABLE',"))
			sbComments.WriteString(fmt.Sprintf("	@level1name=N'%s', ", table.Name))
			sbComments.WriteString(fmt.Sprintln())
			sbComments.WriteString(fmt.Sprintln("	@level2type=N'COLUMN',"))
			sbComments.WriteString(fmt.Sprintf("	@level2name=N'%s' ", column.Name))
			sbComments.WriteString(fmt.Sprintln())
			sbComments.WriteString(fmt.Sprintln("GO"))
		}
	}
	sb.WriteString(fmt.Sprintln(")"))
	sb.WriteString(fmt.Sprintln("GO"))

	if len(primaryKeys) > 0 {
		sb.WriteString(fmt.Sprintln())
		sb.WriteString(fmt.Sprintf("ALTER TABLE [dbo].[%s] ADD CONSTRAINT ", table.Name))
		sb.WriteString(fmt.Sprintln())
		sb.WriteString(fmt.Sprintf("	PK_%s PRIMARY KEY CLUSTERED ", table.Name))
		sb.WriteString(fmt.Sprintf("(%s) ON [PRIMARY] ", strings.Join(primaryKeys, ",")))
		sb.WriteString(fmt.Sprintln())
		sb.WriteString(fmt.Sprintln("GO"))
	}

	sb.WriteString(sbDefaults.String())
	sb.WriteString(sbComments.String())

	if len(table.Description) > 0 {
		sb.WriteString(fmt.Sprintln())
		sb.WriteString(fmt.Sprintln("EXEC [sys].[sp_addextendedproperty]"))
		sb.WriteString(fmt.Sprintln("	@name=N'MS_Description', "))
		sb.WriteString(fmt.Sprintf("	@value=N'%s'", table.Description))
		sb.WriteString(fmt.Sprintln(", "))
		sb.WriteString(fmt.Sprintln("	@level0type=N'SCHEMA',"))
		sb.WriteString(fmt.Sprintln("	@level0name=N'dbo',"))
		sb.WriteString(fmt.Sprintln("	@level1type=N'TABLE',"))
		sb.WriteString(fmt.Sprintf("	@level1name=N'%s'", table.Name))
		sb.WriteString(fmt.Sprintln())
		sb.WriteString(fmt.Sprintln("GO"))
	}
	sb.WriteString(fmt.Sprintln())

	return sb.String(), nil
}

func (s *mssql) ViewDefinition(viewName string) (string, error) {
	db, err := sql.Open(s.connection.DriverName(), s.connection.SourceName())
	if err != nil {
		return "", err
	}
	defer db.Close()

	sb := &strings.Builder{}
	sb.WriteString("select [definition] ")
	sb.WriteString("from [sys].[sql_modules] ")
	sb.WriteString(fmt.Sprintf("where [object_id] = OBJECT_ID('%s') ", viewName))

	query := sb.String()
	row := db.QueryRow(query)

	definition := ""
	err = row.Scan(&definition)
	if err != nil {
		return "", err
	}

	asPos := strings.Index(definition, "AS")
	if asPos > 0 {
		definition = definition[asPos+2:]
	}

	define := &strings.Builder{}
	define.WriteString(fmt.Sprintf("IF OBJECT_ID('[dbo].[%s]') IS NOT NULL ", viewName))
	define.WriteString(fmt.Sprintln(""))
	define.WriteString(fmt.Sprintf("DROP VIEW [dbo].[%s] ", viewName))
	define.WriteString(fmt.Sprintln(""))
	define.WriteString(fmt.Sprintln("GO"))
	define.WriteString(fmt.Sprintf("CREATE VIEW [dbo].[%s] AS ", viewName))
	define.WriteString(definition)

	return define.String(), nil
}

func (s *mssql) columnTypeName(dataType string, length int, precision, scale *int) string {
	sb := &strings.Builder{}
	sb.WriteString(dataType)
	if scale == nil && precision != nil {
		if *precision == 0 {
			sb.WriteString(fmt.Sprintf("(%d)", length))
		}
	} else {
		if strings.ToLower(dataType) == "decimal" || strings.ToLower(dataType) == "numeric" {
			sb.WriteString(fmt.Sprintf("(%d, %d)", precision, *scale))
		}
	}

	return sb.String()
}

func (s *mssql) columnDateDefault(value string) string {
	if strings.HasPrefix(value, "((") && strings.HasSuffix(value, "))") {
		return value[2 : len(value)-2]
	} else if strings.HasPrefix(value, "(N'") && strings.HasSuffix(value, "')") {
		return value[3 : len(value)-2]
	}

	return value
}

func (s *mssql) NewAccess(transactional bool) (sqldb.SqlAccess, error) {
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

func (s *mssql) NewEntity() sqldb.SqlEntity {
	return &entity{}
}

func (s *mssql) NewBuilder() sqldb.SqlBuilder {
	instance := &builder{}
	instance.Reset()

	return instance
}

func (s *mssql) NewFilter(entity interface{}, fieldOr, groupOr bool) sqldb.SqlFilter {
	return newFilter(entity, fieldOr, groupOr)
}

func (s *mssql) IsNoRows(err error) bool {
	if err == nil {
		return false
	}

	if err == sql.ErrNoRows {
		return true
	}

	return false
}

func (s *mssql) Insert(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.Insert(entity)
}

func (s *mssql) InsertSelective(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.InsertSelective(entity)
}

func (s *mssql) Delete(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.Delete(entity, filters...)
}

func (s *mssql) Update(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.Update(entity, filters...)
}

func (s *mssql) UpdateSelective(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.UpdateSelective(entity, filters...)
}

func (s *mssql) UpdateByPrimaryKey(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.UpdateByPrimaryKey(entity)
}

func (s *mssql) UpdateSelectiveByPrimaryKey(entity interface{}) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.UpdateSelectiveByPrimaryKey(entity)
}

func (s *mssql) SelectOne(entity interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectOne(entity, filters...)
}

func (s *mssql) SelectDistinct(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectDistinct(entity, row, order, filters...)
}

func (s *mssql) SelectList(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectList(entity, row, order, filters...)
}

func (s *mssql) SelectPage(entity interface{}, page func(total, page, size, index uint64), row func(index uint64, evt sqldb.SqlEvent), size, index uint64, order interface{}, filters ...sqldb.SqlFilter) error {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectPage(entity, page, row, size, index, order, filters...)
}

func (s *mssql) SelectCount(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlAccess, err := s.NewAccess(false)
	if err != nil {
		return 0, err
	}
	defer sqlAccess.Close()

	return sqlAccess.SelectCount(entity, filters...)
}
