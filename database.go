package querydb

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

//Rows 行
type Rows = sql.Rows

//Result 数据集合
type Result = sql.Result

// Connection 链接
type Connection interface {
	Exec(query string, args ...interface{}) (Result, error)
	Query(query string, args ...interface{}) (*Rows, error)
	NewQuery() *QueryBuilder
	GetLastSql() Sql
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
	config  *Config
	lastsql Sql
}

//QueryTx
type QueryTx struct {
	tx      sql.Tx
	config  *Config
	lastsql Sql
}

//ToMap 将数据接口转化
func ToMap(rows *Rows) []map[string]interface{} {
	cols, err := rows.Columns()
	if err != nil {
		logrus.Println(err)
		return nil
	}
	count := len(cols)
	var data []map[string]interface{}
	vals := make([]string, count)
	ptr := make([]interface{}, count)
	for i := 0; i < count; i++ {
		ptr[i] = &vals[i]
	}
	defer rows.Close()
	for rows.Next() {
		//字段
		entry := make(map[string]interface{}, count)
		err = rows.Scan(ptr...)
		if err != nil {
			logrus.Println(err)
		}
		for i, col := range cols {
			entry[col] = vals[i]
		}
		data = append(data, entry)
	}
	if err = rows.Err(); err != nil {
		logrus.Println(err)
	}
	return data
}

//NewQuery 生成一个新的查询构造器
func (querydb *QueryDb) NewQuery() *QueryBuilder {
	return &QueryBuilder{connection: querydb}
}

//Table 查询构造器快速调用
func (querydb *QueryDb) Table(tablename ...string) *QueryBuilder {
	query := &QueryBuilder{connection: querydb}
	return query.Table(tablename...)
}

//Begin 开启一个事务
func (querydb *QueryDb) Begin() (*QueryTx, error) {
	tx, err := querydb.db.Begin()
	if err != nil {
		return nil, err
	}
	return &QueryTx{tx: *tx, config: querydb.config}, nil
}

//Exec 复用执行语句
func (querydb *QueryDb) Exec(query string, args ...interface{}) (Result, error) {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = args
	start := time.Now()
	defer func() {
		querydb.lastsql.CostTime = time.Since(start)
	}()
	return querydb.db.Exec(query, args...)
}

//Query 复用查询语句
func (querydb *QueryDb) Query(query string, args ...interface{}) (*Rows, error) {
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

//Table 查询构造器快速调用
func (querytx *QueryTx) Table(tablename ...string) *QueryBuilder {
	query := &QueryBuilder{connection: querytx}
	return query.Table(tablename...)
}

//Exec 复用执行语句
func (querytx *QueryTx) Exec(query string, args ...interface{}) (Result, error) {
	querytx.lastsql.Sql = query
	querytx.lastsql.Args = args
	start := time.Now()
	defer func() {
		querytx.lastsql.CostTime = time.Since(start)

	}()
	return querytx.tx.Exec(query, args...)
}

//Query 复用查询语句
func (querytx *QueryTx) Query(query string, args ...interface{}) (*Rows, error) {
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

// ToString sql语句转出string
func (sql Sql) ToString() string {
	s := sql.Sql
	for _, v := range sql.Args {
		val := fmt.Sprintf("%v", v)
		val = strconv.Quote(val)
		s = strings.Replace(s, "?", val, 1)
	}
	return s
}

// ToJson sql语句转出json
func (sql Sql) ToJson() string {
	return fmt.Sprintf(`{"sql":%s,"costtime":"%s"}`, strconv.Quote(sql.ToString()), sql.CostTime)
}
