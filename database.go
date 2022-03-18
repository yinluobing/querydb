package querydb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Connection 链接
type Connection interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	NewQuery() *QueryBuilder
	GetLastSql() Sql
	LastSql(query string, args ...interface{})
}

// Sql sql语句
type Sql struct {
	Sql      string
	Args     []interface{}
	CostTime time.Duration
}

// QueryDb mysql 配置
type QueryDb struct {
	db      *sql.DB
	lastsql Sql
	link    *Config
}

//QueryTx
type QueryTx struct {
	Tx      *sql.Tx
	lastsql Sql
	link    *Config
}

//NewQuery 生成一个新的查询构造器
func (querydb *QueryDb) NewQuery() *QueryBuilder {
	return &QueryBuilder{connection: querydb, debug: querydb.link.Debug}
}

//Begin 开启一个事务
func (querydb *QueryDb) Begin() (*QueryTx, error) {
	tx, err := querydb.db.Begin()
	if err != nil {
		return nil, err
	}
	return &QueryTx{Tx: tx, link: querydb.link}, nil
}

//Exec 复用执行语句
func (querydb *QueryDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = args
	start := time.Now()
	defer func() {
		querydb.lastsql.CostTime = time.Since(start)
	}()
	ctx := context.TODO()
	var res sql.Result
	var err error

	//添加预处理
	stmt, err := querydb.db.PrepareContext(ctx, query)

	if err != nil {
		querydb.db.PingContext(ctx)
		return res, err
	}
	defer stmt.Close()
	res, err = stmt.ExecContext(ctx, args...)
	querydb.db.PingContext(ctx)

	return res, err
}

//Query 复用查询语句
func (querydb *QueryDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = args
	start := time.Now()
	defer func() {
		querydb.lastsql.CostTime = time.Since(start)
	}()
	ctx := context.TODO()
	var res *sql.Rows
	var err error

	//添加预处理
	stmt, err := querydb.db.PrepareContext(ctx, query)
	if err != nil {
		querydb.db.PingContext(ctx)
		return res, err
	}
	defer stmt.Close()
	res, err = stmt.QueryContext(ctx, args...)
	querydb.db.PingContext(ctx)
	return res, err
}

//GetLastSql 获取sql语句
func (querydb *QueryDb) GetLastSql() Sql {
	return querydb.lastsql
}

// Commit 事务提交
func (querytx *QueryTx) Commit() error {
	return querytx.Tx.Commit()
}

// Rollback 事务回滚
func (querytx *QueryTx) Rollback() error {
	return querytx.Tx.Rollback()
}

// NewQuery 生成一个新的查询构造器
func (querytx *QueryTx) NewQuery() *QueryBuilder {
	return &QueryBuilder{connection: querytx}
}

//Exec 复用执行语句
func (querytx *QueryTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	querytx.lastsql.Sql = query
	querytx.lastsql.Args = args
	start := time.Now()
	defer func() {
		querytx.lastsql.CostTime = time.Since(start)

	}()
	ctx := context.TODO()
	var res sql.Result
	var err error
	//添加预处理
	stmt, err := querytx.Tx.PrepareContext(ctx, query)
	if err != nil {
		return res, err
	}
	res, err = stmt.ExecContext(ctx, args...)
	return res, err

}

//Query 复用查询语句
func (querytx *QueryTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	querytx.lastsql.Sql = query
	querytx.lastsql.Args = args
	start := time.Now()
	defer func() {
		querytx.lastsql.CostTime = time.Since(start)
	}()
	ctx := context.TODO()
	var res *sql.Rows
	var err error

	//添加预处理
	stmt, err := querytx.Tx.PrepareContext(ctx, query)
	if err != nil {
		return res, err
	}
	res, err = stmt.QueryContext(ctx, args...)
	return res, err
}

//GetLastSql 获取sql语句
func (querytx *QueryTx) GetLastSql() Sql {
	return querytx.lastsql
}

func (querytx *QueryTx) LastSql(query string, args ...interface{}) {
	querytx.lastsql.Sql = query
	querytx.lastsql.Args = args
}

func (querydb *QueryDb) LastSql(query string, args ...interface{}) {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = args
}

// ToString sql语句转出string
func (sqlRaw Sql) ToString() string {
	s := sqlRaw.Sql
	for _, v := range sqlRaw.Args {
		if isNilFixed(v) {
			v = "NULL"
		} else {
			switch reflect.ValueOf(v).Interface().(type) {
			case sql.NullString:
				v = sqlRaw.nullString(v.(sql.NullString))
			case sql.NullInt64:
				v = sqlRaw.nullInt64(v.(sql.NullInt64))
			case sql.NullInt32:
				v = sqlRaw.nullInt32(v.(sql.NullInt32))
			case sql.NullFloat64:
				v = sqlRaw.nullFloat64(v.(sql.NullFloat64))
			case sql.NullBool:
				v = sqlRaw.nullBool(v.(sql.NullBool))
			case sql.NullTime:
				v = sqlRaw.nullTime(v.(sql.NullTime))
			}
		}
		s = convert(s, v)
	}
	return s
}

func isNilFixed(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func convert(s string, v interface{}) string {
	switch reflect.ValueOf(v).Interface().(type) {
	case string:
		if val := fmt.Sprintf("%v", v); val == "NULL" {
			return strings.Replace(s, "?", fmt.Sprintf("%v", v), 1)
		} else {
			return strings.Replace(s, "?", strconv.Quote(fmt.Sprintf("%v", v)), 1)
		}
	}
	return strings.Replace(s, "?", fmt.Sprintf("%v", v), 1)
}

func (sqlRaw Sql) nullTime(s sql.NullTime) interface{} {
	if s.Valid {
		return s.Time.Unix()
	}
	return "NULL"
}

func (sqlRaw Sql) nullBool(s sql.NullBool) interface{} {
	if s.Valid {
		if s.Bool {
			return 1
		} else {
			return 0
		}
	}
	return "NULL"
}

func (sqlRaw Sql) nullFloat64(s sql.NullFloat64) interface{} {
	if s.Valid {
		return s.Float64
	}
	return "NULL"
}

func (sqlRaw Sql) nullInt32(s sql.NullInt32) interface{} {
	if s.Valid {
		return s.Int32
	}
	return "NULL"
}

func (sqlRaw Sql) nullInt64(s sql.NullInt64) interface{} {
	if s.Valid {
		return s.Int64
	}
	return "NULL"
}
func (sqlRaw Sql) nullString(s sql.NullString) interface{} {
	if s.Valid {
		return s.String
	}
	return "NULL"
}

// ToJson sql语句转出json
func (sql Sql) ToJson() string {
	return fmt.Sprintf(`{"sql":%s,"costtime":"%s"}`, strconv.Quote(sql.ToString()), sql.CostTime)
}
