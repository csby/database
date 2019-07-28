package oracle

import (
	"database/sql"
	"github.com/csby/database/sqldb"
)

type normal struct {
	access

	db *sql.DB
}

func (s *normal) Close() error {
	return s.db.Close()
}

func (s *normal) Commit() error {
	return nil
}

func (s *normal) Version() int {
	return 0
}

func (s *normal) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *normal) Prepare(query string) (*sql.Stmt, error) {
	return s.db.Prepare(query)
}

func (s *normal) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

func (s *normal) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.db.QueryRow(query, args...)
}

func (s *normal) IsNoRows(err error) bool {
	return s.isNoRows(err)
}

func (s *normal) Insert(entity interface{}) (uint64, error) {
	return s.insert(s, false, entity)
}

func (s *normal) InsertSelective(entity interface{}) (uint64, error) {
	return s.insert(s, true, entity)
}

func (s *normal) Delete(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	return s.delete(s, entity, filters...)
}

func (s *normal) Update(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	return s.update(s, false, entity, filters...)
}

func (s *normal) UpdateSelective(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	return s.update(s, true, entity, filters...)
}

func (s *normal) UpdateByPrimaryKey(entity interface{}) (uint64, error) {
	return s.updateByPrimaryKey(s, false, entity)
}

func (s *normal) UpdateSelectiveByPrimaryKey(entity interface{}) (uint64, error) {
	return s.updateByPrimaryKey(s, true, entity)
}

func (s *normal) SelectOne(entity interface{}, filters ...sqldb.SqlFilter) error {
	return s.selectOne(s, entity, filters...)
}

func (s *normal) SelectDistinct(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	return s.selectList(s, true, entity, row, order, filters...)
}

func (s *normal) SelectList(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	return s.selectList(s, false, entity, row, order, filters...)
}

func (s *normal) SelectPage(entity interface{}, page func(total, page, size, index uint64), row func(index uint64, evt sqldb.SqlEvent), size, index uint64, order interface{}, filters ...sqldb.SqlFilter) error {
	return s.selectPage(s, entity, page, row, size, index, order, filters...)
}

func (s *normal) SelectCount(dbEntity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlEntity := &entity{}
	err := sqlEntity.Parse(dbEntity)
	if err != nil {
		return 0, err
	}

	return s.selectCount(s, sqlEntity.Name(), filters...)
}
