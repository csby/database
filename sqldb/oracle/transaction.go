package oracle

import (
	"context"
	"database/sql"
	"github.com/csby/database/sqldb"
)

type transaction struct {
	access

	db *sql.DB
	tx *sql.Tx
}

func (s *transaction) Close() error {
	defer s.db.Close()

	return s.tx.Rollback()
}

func (s *transaction) Commit() error {
	return s.tx.Commit()
}

func (s *transaction) Rollback() error {
	return s.tx.Rollback()
}

func (s *transaction) Version() int {
	return 0
}

func (s *transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.tx.Exec(query, args...)
}

func (s *transaction) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.tx.ExecContext(ctx, query, args...)
}

func (s *transaction) Prepare(query string) (*sql.Stmt, error) {
	return s.tx.Prepare(query)
}

func (s *transaction) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return s.tx.PrepareContext(ctx, query)
}

func (s *transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.tx.Query(query, args...)
}

func (s *transaction) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.tx.QueryContext(ctx, query, args...)
}

func (s *transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.tx.QueryRow(query, args...)
}

func (s *transaction) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.tx.QueryRowContext(ctx, query, args)
}

func (s *transaction) Stmt(stmt *sql.Stmt) *sql.Stmt {
	return s.tx.Stmt(stmt)
}

func (s *transaction) StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt {
	return s.tx.StmtContext(ctx, stmt)
}

func (s *transaction) IsNoRows(err error) bool {
	return s.isNoRows(err)
}

func (s *transaction) Insert(entity interface{}, fields ...sqldb.SqlField) (uint64, error) {
	return s.insert(s, false, entity)
}

func (s *transaction) InsertSelective(entity interface{}) (uint64, error) {
	return s.insert(s, true, entity)
}

func (s *transaction) Delete(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	return s.delete(s, entity, filters...)
}

func (s *transaction) Update(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	return s.update(s, false, entity, filters...)
}

func (s *transaction) UpdateSelective(entity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	return s.update(s, true, entity, filters...)
}

func (s *transaction) UpdateByPrimaryKey(entity interface{}) (uint64, error) {
	return s.updateByPrimaryKey(s, false, entity)
}

func (s *transaction) UpdateSelectiveByPrimaryKey(entity interface{}) (uint64, error) {
	return s.updateByPrimaryKey(s, true, entity)
}

func (s *transaction) SelectOne(entity interface{}, filters ...sqldb.SqlFilter) error {
	return s.selectOne(s, entity, filters...)
}

func (s *transaction) SelectDistinct(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	return s.selectList(s, true, entity, row, order, filters...)
}

func (s *transaction) SelectList(entity interface{}, row func(index uint64, evt sqldb.SqlEvent), order interface{}, filters ...sqldb.SqlFilter) error {
	return s.selectList(s, false, entity, row, order, filters...)
}

func (s *transaction) SelectPage(entity interface{}, page func(total, page, size, index uint64), row func(index uint64, evt sqldb.SqlEvent), size, index uint64, order interface{}, filters ...sqldb.SqlFilter) error {
	return s.selectPage(s, entity, page, row, size, index, order, filters...)
}

func (s *transaction) SelectCount(dbEntity interface{}, filters ...sqldb.SqlFilter) (uint64, error) {
	sqlEntity := &entity{}
	err := sqlEntity.Parse(dbEntity)
	if err != nil {
		return 0, err
	}

	return s.selectCount(s, sqlEntity.Name(), filters...)
}
