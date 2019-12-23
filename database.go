package querydb

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//Rows 行
// type Rows = sql.Rows

//Result 数据集合
// type Result = sql.Result

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
	// ctx context.Context
	db *sql.DB
	// config  *Config
	lastsql Sql
}

//QueryTx
type QueryTx struct {
	// ctx context.Context
	tx sql.Tx
	// config  *Config
	lastsql Sql
}

//NewQuery 生成一个新的查询构造器
func (querydb *QueryDb) NewQuery() *QueryBuilder {
	return &QueryBuilder{connection: querydb}
}

//Begin 开启一个事务
func (querydb *QueryDb) Begin() (*QueryTx, error) {
	tx, err := querydb.db.Begin()
	if err != nil {
		return nil, err
	}
	return &QueryTx{tx: *tx}, nil
}

//Exec 复用执行语句
func (querydb *QueryDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = args
	start := time.Now()
	defer func() {
		querydb.lastsql.CostTime = time.Since(start)
	}()
	return querydb.db.Exec(query, args...)
}

//Query 复用查询语句
func (querydb *QueryDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = args
	start := time.Now()
	defer func() {
		querydb.lastsql.CostTime = time.Since(start)
	}()
	return querydb.db.Query(query, args...)
}

//GetLastSql 获取sql语句
func (querydb *QueryDb) GetLastSql() Sql {
	return querydb.lastsql
}

// Commit 事务提交
func (querytx *QueryTx) Commit() error {
	return querytx.tx.Commit()
}

// Rollback 事务回滚
func (querytx *QueryTx) Rollback() error {
	return querytx.tx.Rollback()
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
	return querytx.tx.Exec(query, args...)
}

//Query 复用查询语句
func (querytx *QueryTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	querytx.lastsql.Sql = query
	querytx.lastsql.Args = args
	start := time.Now()
	defer func() {
		querytx.lastsql.CostTime = time.Since(start)
	}()
	return querytx.tx.Query(query, args...)
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
		switch reflect.ValueOf(v).Interface().(type) {
		case sql.NullString:
			v = sqlRaw.nullString(v.(sql.NullString))
			// case sql.NullTime:
			// 	v = sqlRaw.nullTime(v.(sql.NullTime))
		}
		val := fmt.Sprintf("%v", v)
		val = strconv.Quote(val)
		s = strings.Replace(s, "?", val, 1)
	}
	return s
}

func (sqlRaw Sql) nullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

// ToJson sql语句转出json
func (sql Sql) ToJson() string {
	return fmt.Sprintf(`{"sql":%s,"costtime":"%s"}`, strconv.Quote(sql.ToString()), sql.CostTime)
}
