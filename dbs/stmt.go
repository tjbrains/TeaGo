package dbs

import (
	"database/sql"
	"iter"

	"github.com/tjbrains/TeaGo/maps"
)

// Stmt SQL语句
type Stmt struct {
	rawStmt *sql.Stmt
}

// NewStmt 构造
func NewStmt(stmt *sql.Stmt) *Stmt {
	return &Stmt{
		rawStmt: stmt,
	}
}

func (this *Stmt) Query(args ...any) (*sql.Rows, error) {
	return this.rawStmt.Query(args...)
}

func (this *Stmt) FindRows(args ...any) (rows *Rows, err error) {
	rawRows, err := this.rawStmt.Query(args...)
	if err != nil {
		return nil, err
	}

	rows = NewRows(rawRows)
	return
}

func (this *Stmt) FindOnes(args ...any) (ones []maps.Map, columnNames []string, err error) {
	rawRows, err := this.rawStmt.Query(args...)
	if err != nil {
		return nil, nil, err
	}

	var rows = NewRows(rawRows)
	defer func() {
		_ = rows.Close()
	}()

	columnNames, err = rows.Columns()
	if err != nil {
		return
	}

	ones, err = rows.FindOnes()
	return
}

func (this *Stmt) FindOnesSeq(args ...any) (seq iter.Seq[maps.Map], columnNames []string, err error) {
	rawRows, err := this.rawStmt.Query(args...)
	if err != nil {
		return nil, nil, err
	}

	var rows = NewRows(rawRows)

	columnNames, err = rows.Columns()
	if err != nil {
		_ = rows.Close()
		return
	}

	rowsSeq, err := rows.FindOnesSeq()
	if err != nil {
		_ = rows.Close()
		return
	}

	seq = iter.Seq[maps.Map](func(yield func(v maps.Map) bool) {
		defer func() {
			_ = rows.Close()
		}()

		for v := range rowsSeq {
			if !yield(v) {
				return
			}
		}
	})

	return
}

func (this *Stmt) FindOne(args ...any) (one maps.Map, err error) {
	rawRows, err := this.rawStmt.Query(args...)
	if err != nil {
		return nil, err
	}

	var rows = NewRows(rawRows)
	defer func() {
		_ = rows.Close()
	}()

	return rows.FindOne()
}

func (this *Stmt) FindCol(colIndex int, args ...any) (colValue any, err error) {
	rawRows, err := this.rawStmt.Query(args...)
	if err != nil {
		return nil, err
	}

	var rows = NewRows(rawRows)
	defer func() {
		_ = rows.Close()
	}()

	return rows.FindCol(colIndex)
}

func (this *Stmt) Exec(args ...any) (sql.Result, error) {
	return this.rawStmt.Exec(args...)
}

// Close 关闭
func (this *Stmt) Close() error {
	return this.rawStmt.Close()
}

// Raw 获取原始的语句
func (this *Stmt) Raw() *sql.Stmt {
	return this.rawStmt
}
