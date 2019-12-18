package querydb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

// Sql sql语句
type Sql struct {
	Sql      string
	Args     []interface{}
	CostTime time.Duration
}

// QueryDb mysql 配置
type QueryDb struct {
	ctx     context.Context
	db      *sql.DB
	tx      *sql.Tx
	conn    *sql.Conn
	lastsql Sql
}

//WithContext 上下文
func (querydb *QueryDb) WithContext(ctx context.Context) *QueryDb {
	querydb.ctx = ctx
	return querydb
}

//Select 查询
func (querydb *QueryDb) Select(query string, bindings []interface{}) *Rows {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = bindings
	start := time.Now()
	defer func() {
		querydb.lastsql.CostTime = time.Since(start)
	}()
	rows, err := querydb.queryRows(query, bindings)
	if err != nil {
		err = NewDBError(err.Error(), querydb.GetLastSQL())
		logrus.Println(err.Error())
		return &Rows{rs: nil, lastError: err}
	}
	return &Rows{rs: rows, lastError: err}
}

//Insert 数据插入
func (querydb *QueryDb) Insert(query string, bindings []interface{}) (int64, error) {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = bindings
	start := time.Now()
	defer func() {
		querydb.lastsql.CostTime = time.Since(start)
	}()
	rs, err := querydb.exec(query, bindings)
	if err != nil {
		err = NewDBError(err.Error(), querydb.GetLastSQL())
		logrus.Println(err.Error())
		return 0, err
	}
	return rs.LastInsertId()
}

//MultiInsert 多条数据插入
func (querydb *QueryDb) MultiInsert(query string, bindingsArr [][]interface{}) ([]int64, error) {
	var stmt *sql.Stmt
	var err error
	if querydb.tx != nil {
		stmt, err = querydb.tx.PrepareContext(querydb.ctx, query)
	} else {
		var conn *sql.Conn
		conn, err = querydb.getConn()
		if err != nil {
			logrus.Println(err.Error())
			return nil, err
		}
		stmt, err = conn.PrepareContext(querydb.ctx, query)
	}
	if err != nil {
		logrus.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()
	lastInsertIds := make([]int64, 0)
	for _, bindings := range bindingsArr {
		querydb.lastsql.Sql = query
		querydb.lastsql.Args = bindings
		start := time.Now()
		defer func() {
			querydb.lastsql.CostTime = time.Since(start)
		}()
		rs, err := stmt.ExecContext(querydb.ctx, bindings...)

		if err != nil {
			err = NewDBError(err.Error(), querydb.GetLastSQL())
			logrus.Println(err.Error())
			return nil, err
		}
		lastInsertID, err := rs.LastInsertId()
		if err != nil {
			err = NewDBError(err.Error(), querydb.GetLastSQL())
			logrus.Println(err.Error())
			return nil, err
		}
		lastInsertIds = append(lastInsertIds, lastInsertID)
	}
	return lastInsertIds, nil
}

//Update 更新
func (querydb *QueryDb) Update(query string, bindings []interface{}) (int64, error) {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = bindings
	start := time.Now()
	defer func() {
		querydb.lastsql.CostTime = time.Since(start)
	}()
	rs, err := querydb.exec(query, bindings)
	if err != nil {
		err = NewDBError(err.Error(), querydb.GetLastSQL())
		logrus.Println(err.Error())
		return 0, err
	}
	return rs.RowsAffected()
}

//Delete 更新
func (querydb *QueryDb) Delete(query string, bindings []interface{}) (int64, error) {
	querydb.lastsql.Sql = query
	querydb.lastsql.Args = bindings
	start := time.Now()
	defer func() {
		querydb.lastsql.CostTime = time.Since(start)
	}()
	rs, err := querydb.exec(query, bindings)
	if err != nil {
		err = NewDBError(err.Error(), querydb.GetLastSQL())
		logrus.Println(err.Error())
		return 0, err
	}

	return rs.RowsAffected()
}

//Begin 事务
func (querydb *QueryDb) Begin() error {
	if querydb.tx == nil {
		conn, err := querydb.getConn()
		if err != nil {
			return err
		}
		tx, err := conn.BeginTx(querydb.ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			logrus.Println(err.Error())
			return err
		}
		querydb.tx = tx
	}
	return nil
}

//Commit 事务提交
func (querydb *QueryDb) Commit() error {
	if querydb.tx == nil {
		return errors.New("no beginTx")
	}
	return querydb.tx.Commit()
}

//RollBack 事务回滚
func (querydb *QueryDb) RollBack() error {
	if querydb.tx == nil {
		return errors.New("no beginTx")
	}
	return querydb.tx.Rollback()
}
func (querydb *QueryDb) queryRows(query string, bindings []interface{}) (rows *sql.Rows, err error) {
	if querydb.tx != nil {
		rows, err = querydb.tx.QueryContext(querydb.ctx, query, bindings...)
		return
	}
	var conn *sql.Conn
	conn, err = querydb.getConn()
	if err != nil {
		logrus.Println(err.Error())
		return nil, err
	}
	rows, err = conn.QueryContext(querydb.ctx, query, bindings...)
	return
}

func (querydb *QueryDb) getConn() (*sql.Conn, error) {
	var err error
	var db *sql.DB = querydb.db
	if querydb.conn != nil {
		return querydb.conn, nil
	}
	conn, err := db.Conn(querydb.ctx)
	if err != nil {
		logrus.Println(err.Error())
		return nil, err
	}
	querydb.conn = conn
	return querydb.conn, nil
}

func (querydb *QueryDb) exec(query string, bindings []interface{}) (rs sql.Result, err error) {
	if querydb.tx != nil {
		rs, err = querydb.tx.ExecContext(querydb.ctx, query, bindings...)
		return
	}
	var conn *sql.Conn
	conn, err = querydb.getConn()
	if err != nil {
		logrus.Println(err.Error())
		return nil, err
	}
	rs, err = conn.ExecContext(querydb.ctx, query, bindings...)
	return
}

//GetLastSQL 获取sql语句
func (querydb *QueryDb) GetLastSQL() Sql {
	return querydb.lastsql
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

// ToJSON sql语句转出json
func (sql Sql) ToJSON() string {
	return fmt.Sprintf(`{"sql":%s,"costtime":"%s"}`, strconv.Quote(sql.ToString()), sql.CostTime)
}

//Tabel 表名
func (querydb *QueryDb) Table(table string) *Builder {
	return querydb.query().Table(table)
}

//query 构建查询语句
func (querydb *QueryDb) query() *Builder {
	g := NewGrammar()
	b := newBuilder(querydb, g)
	return b
}
