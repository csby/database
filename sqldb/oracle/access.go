package oracle

import (
	"database/sql"
	"fmt"
	"github.com/csby/database/sqldb"
	"strings"
)

type access struct {
}

func (s *access) NewFilter(entity interface{}, fieldOr, groupOr bool) sqldb.SqlFilter {
	return newFilter(entity, fieldOr, groupOr)
}

func (s *access) isNoRows(err error) bool {
	if err == nil {
		return false
	}

	if err == sql.ErrNoRows {
		return true
	}

	return false
}

func (s *access) getFilterFields(dbFilter interface{}) []sqldb.SqlField {
	fields := make([]sqldb.SqlField, 0)
	if dbFilter == nil {
		return fields
	}

	filterEntity := &entity{}
	err := filterEntity.ParseFilter(dbFilter)
	if err != nil {
		return fields
	}
	fieldCount := filterEntity.FieldCount()
	for fieldIndex := 0; fieldIndex < fieldCount; fieldIndex++ {
		field := filterEntity.Field(fieldIndex)
		if field.ValueEmpty() {
			continue
		}
		fields = append(fields, field)
	}

	return fields
}

func (s *access) fillWhereField(sqlBuilder sqldb.SqlBuilder, fields []sqldb.SqlField, or bool) {
	if sqlBuilder == nil {
		return
	}

	fieldCount := len(fields)
	if fieldCount > 0 {
		sqlBuilder.AppendFormat("(")
		for fieldIndex := 0; fieldIndex < fieldCount; fieldIndex++ {
			field := fields[fieldIndex]
			filterSymbol := field.Filter()

			if strings.ToLower(filterSymbol) == "in" {
				if fieldIndex == 0 {
					sqlBuilder.WhereFormat("%s %s %s", field.Name(), filterSymbol, field.Value())
				} else if or {
					sqlBuilder.WhereFormatOr("%s %s %s", field.Name(), filterSymbol, field.Value())
				} else {
					sqlBuilder.WhereFormatAnd("%s %s %s", field.Name(), filterSymbol, field.Value())
				}
			} else {
				if fieldIndex == 0 {
					sqlBuilder.Where(fmt.Sprintf("%s %s %s", field.Name(), filterSymbol, sqlBuilder.ArgName()), field.Value())
				} else if or {
					sqlBuilder.WhereOr(fmt.Sprintf("%s %s %s", field.Name(), filterSymbol, sqlBuilder.ArgName()), field.Value())
				} else {
					sqlBuilder.WhereAnd(fmt.Sprintf("%s %s %s", field.Name(), filterSymbol, sqlBuilder.ArgName()), field.Value())
				}
			}
		}
		sqlBuilder.AppendFormat(")")
	}
}

func (s *access) fillWhereFilter(sqlBuilder sqldb.SqlBuilder, filters []sqldb.SqlFilter) {
	filterCount := len(filters)
	if filterCount < 1 {
		return
	}

	for filterIndex := 0; filterIndex < filterCount; filterIndex++ {
		filter := filters[filterIndex]
		fields := s.getFilterFields(filter.Fields())
		if len(fields) < 1 {
			continue
		}

		if filter.GroupOr() {
			sqlBuilder.WhereOr("")
		} else {
			sqlBuilder.WhereAnd("")
		}

		s.fillWhereField(sqlBuilder, fields, filter.FieldOr())
	}
}

func (s *access) fillWhere(sqlBuilder sqldb.SqlBuilder, filters ...sqldb.SqlFilter) {
	s.fillWhereFilter(sqlBuilder, filters)
}

func (s *access) fillOrder(sqlBuilder sqldb.SqlBuilder, order interface{}) {
	if order == nil {
		return
	}
	sqlEntity := &entity{}
	err := sqlEntity.Parse(order)
	if err != nil {
		return
	}

	count := len(sqlEntity.fields)
	if count < 1 {
		return
	}
	sqlBuilder.Append(fmt.Sprintf("order by %s %s", sqlEntity.fields[0].name, sqlEntity.fields[0].order))

	for i := 1; i < count; i++ {
		sqlBuilder.Append(fmt.Sprintf(", %s %s", sqlEntity.fields[i].name, sqlEntity.fields[i].order))
	}
}

func (s *access) insert(sqlAccess sqldb.SqlAccess, selective bool, dbEntity interface{}) (uint64, error) {
	sqlEntity := &entity{}
	err := sqlEntity.Parse(dbEntity)
	if err != nil {
		return 0, err
	}

	hasAutoField := false
	sqlBuilder := &builder{}
	sqlBuilder.Reset()
	sqlBuilder.Insert(sqlEntity.Name())
	fieldCount := sqlEntity.FieldCount()
	for fieldIndex := 0; fieldIndex < fieldCount; fieldIndex++ {
		field := sqlEntity.Field(fieldIndex)
		if field.AutoIncrement() {
			hasAutoField = true
			continue
		}
		if selective {
			if field.ValueEmpty() {
				continue
			}
		}

		sqlBuilder.Value(field.Name(), field.Value())
	}

	stmt, err := sqlAccess.Prepare(sqlBuilder.Query())
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(sqlBuilder.Args()...)
	if err != nil {
		return 0, err
	}

	if hasAutoField {
		id, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		return uint64(id), nil
	}

	return 0, nil
}

func (s *access) delete(sqlAccess sqldb.SqlAccess, dbEntity interface{}, sqlFilters ...sqldb.SqlFilter) (uint64, error) {
	sqlEntity := &entity{}
	err := sqlEntity.Parse(dbEntity)
	if err != nil {
		return 0, err
	}

	sqlBuilder := &builder{}
	sqlBuilder.Reset()
	sqlBuilder.Delete(sqlEntity.Name())
	s.fillWhere(sqlBuilder, sqlFilters...)

	stmt, err := sqlAccess.Prepare(sqlBuilder.Query())
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(sqlBuilder.Args()...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return uint64(rowsAffected), nil
}

func (s *access) update(sqlAccess sqldb.SqlAccess, selective bool, dbEntity interface{}, sqlFilters ...sqldb.SqlFilter) (uint64, error) {
	sqlEntity := &entity{}
	err := sqlEntity.Parse(dbEntity)
	if err != nil {
		return 0, err
	}

	sqlBuilder := &builder{}
	sqlBuilder.Reset()
	sqlBuilder.Update(sqlEntity.Name())
	fieldCount := sqlEntity.FieldCount()
	for fieldIndex := 0; fieldIndex < fieldCount; fieldIndex++ {
		field := sqlEntity.Field(fieldIndex)
		if field.AutoIncrement() {
			continue
		}
		if selective {
			if field.ValueEmpty() {
				continue
			}
		}

		sqlBuilder.Set(field.Name(), field.Value())
	}
	s.fillWhere(sqlBuilder, sqlFilters...)

	stmt, err := sqlAccess.Prepare(sqlBuilder.Query())
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(sqlBuilder.Args()...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return uint64(rowsAffected), nil
}

func (s *access) updateByPrimaryKey(sqlAccess sqldb.SqlAccess, selective bool, dbEntity interface{}) (uint64, error) {
	sqlEntity := &entity{}
	err := sqlEntity.Parse(dbEntity)
	if err != nil {
		return 0, err
	}

	sqlBuilder := &builder{}
	sqlBuilder.Reset()
	sqlBuilder.Update(sqlEntity.Name())
	fieldCount := sqlEntity.FieldCount()
	primaryFields := make([]sqldb.SqlField, 0)
	for fieldIndex := 0; fieldIndex < fieldCount; fieldIndex++ {
		field := sqlEntity.Field(fieldIndex)
		if field.PrimaryKey() {
			primaryFields = append(primaryFields, field)
			continue
		}
		if field.AutoIncrement() {
			continue
		}
		if selective {
			if field.ValueEmpty() {
				continue
			}
		}

		sqlBuilder.Set(field.Name(), field.Value())
	}

	primaryCount := len(primaryFields)
	if primaryCount < 1 {
		return 0, fmt.Errorf("no primary key")
	}
	for fieldIndex := 0; fieldIndex < primaryCount; fieldIndex++ {
		field := primaryFields[fieldIndex]
		sqlBuilder.Where(fmt.Sprintf(" %s=?", field.Name()), field.Value())
	}

	query := sqlBuilder.Query()
	stmt, err := sqlAccess.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	args := sqlBuilder.Args()
	result, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rowsAffected == 0 {
		sqlBuilder.Reset()
		sqlBuilder.Select("COUNT(*)", false).From(sqlEntity.Name())
		for fieldIndex := 0; fieldIndex < primaryCount; fieldIndex++ {
			field := primaryFields[fieldIndex]
			sqlBuilder.Where(fmt.Sprintf(" %s=?", field.Name()), field.Value())
		}

		query := sqlBuilder.Query()
		row := sqlAccess.QueryRow(query, sqlBuilder.Args()...)
		err := row.Scan(&rowsAffected)
		if err != nil {
			return 0, err
		}
	}

	return uint64(rowsAffected), nil
}

func (s *access) selectCount(sqlAccess sqldb.SqlAccess, tableName string, sqlFilters ...sqldb.SqlFilter) (uint64, error) {
	sqlBuilder := &builder{}
	sqlBuilder.Reset()
	sqlBuilder.Select("COUNT(*)", false).From(tableName)
	s.fillWhere(sqlBuilder, sqlFilters...)

	count := uint64(0)
	query := sqlBuilder.Query()
	row := sqlAccess.QueryRow(query, sqlBuilder.Args()...)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *access) selectOne(sqlAccess sqldb.SqlAccess, dbEntity interface{}, sqlFilters ...sqldb.SqlFilter) error {
	sqlEntity := &entity{}
	err := sqlEntity.Parse(dbEntity)
	if err != nil {
		return err
	}

	sqlBuilder := &builder{}
	sqlBuilder.Reset()
	sqlBuilder.Select(sqlEntity.ScanFields(), false).From(sqlEntity.Name())
	s.fillWhere(sqlBuilder, sqlFilters...)

	query := sqlBuilder.Query()
	row := sqlAccess.QueryRow(query, sqlBuilder.Args()...)
	err = row.Scan(sqlEntity.ScanArgs()...)
	if err != nil {
		return err
	}

	return nil
}

func (s *access) selectList(sqlAccess sqldb.SqlAccess, distinct bool, dbEntity interface{}, row func(index uint64, evt sqldb.SqlEvent), dbOrder interface{}, sqlFilters ...sqldb.SqlFilter) error {
	sqlEntity := &entity{}
	err := sqlEntity.Parse(dbEntity)
	if err != nil {
		return err
	}

	sqlBuilder := &builder{}
	sqlBuilder.Reset()
	sqlBuilder.Select(sqlEntity.ScanFields(), distinct).From(sqlEntity.Name())
	s.fillWhere(sqlBuilder, sqlFilters...)
	s.fillOrder(sqlBuilder, dbOrder)

	query := sqlBuilder.Query()
	args := sqlBuilder.Args()
	rows, err := sqlAccess.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	idx := uint64(0)
	evt := &event{canceled: false, err: nil}
	for rows.Next() {
		err = rows.Scan(sqlEntity.ScanArgs()...)
		if err != nil {
			return err
		}

		if row != nil {
			row(idx, evt)
			idx++
		}

		if evt.canceled {
			return evt.err
		}
	}

	return nil
}

func (s *access) selectPage(sqlAccess sqldb.SqlAccess, dbEntity interface{}, page func(total, page, size, index uint64), row func(index uint64, evt sqldb.SqlEvent), size, index uint64, dbOrder interface{}, sqlFilters ...sqldb.SqlFilter) error {
	sqlEntity := &entity{}
	err := sqlEntity.Parse(dbEntity)
	if err != nil {
		return err
	}
	total, err := s.selectCount(sqlAccess, sqlEntity.Name(), sqlFilters...)
	if err != nil {
		return err
	}
	if size < 1 {
		size = 1
	}
	pageCount := total / size
	if (total % size) != 0 {
		pageCount++
	}
	pageIndex := index
	if pageIndex > pageCount {
		pageIndex = pageCount
	} else if pageIndex < 1 {
		pageIndex = 1
	}
	if page != nil {
		page(total, pageCount, size, pageIndex)
	}
	if total < 1 {
		return nil
	}

	sqlBuilder := &builder{}
	sqlBuilder.Reset()
	sqlBuilder.Select(sqlEntity.ScanFields(), false).From(sqlEntity.Name())
	s.fillWhere(sqlBuilder, sqlFilters...)
	s.fillOrder(sqlBuilder, dbOrder)

	startIndex := (pageIndex - 1) * size
	sqlBuilder.Append("LIMIT ?, ?", startIndex, size)

	query := sqlBuilder.Query()
	args := sqlBuilder.Args()
	rows, err := sqlAccess.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	idx := uint64(0)
	evt := &event{canceled: false, err: nil}
	for rows.Next() {
		err = rows.Scan(sqlEntity.ScanArgs()...)
		if err != nil {
			return err
		}

		if row != nil {
			row(idx, evt)
			idx++
		}

		if evt.canceled {
			return evt.err
		}
	}

	return nil
}
