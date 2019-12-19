package main

import (
	"database/sql"
	"fmt"

	"github.com/pm-esd/querydb"
)

func main() {
	master := &querydb.Config{
		Username:     "root",
		Password:     "mysql",
		Host:         "127.0.0.1",
		Port:         "3306",
		Charset:      "utf8mb4",
		Database:     "ott",
		MaxLifetime:  120,
		MaxIdleConns: 200,
		MaxOpenConns: 800,
	}
	slave1 := &querydb.Config{
		Username:     "root",
		Password:     "mysql",
		Host:         "127.0.0.1",
		Port:         "3306",
		Charset:      "utf8mb4",
		Database:     "ott",
		MaxLifetime:  120,
		MaxIdleConns: 200,
		MaxOpenConns: 800,
	}
	slave2 := &querydb.Config{
		Username:     "root",
		Password:     "mysql",
		Host:         "127.0.0.1",
		Port:         "3306",
		Charset:      "utf8mb4",
		Database:     "ott",
		MaxLifetime:  120,
		MaxIdleConns: 200,
		MaxOpenConns: 800,
	}
	master.SetSlave(slave1)
	master.SetSlave(slave2)
	instance := querydb.Default()
	instance.SetConfig("test", master)
	db := instance.Write("test") //主库

	type user struct {
		Id   int            `json:"id"` //tag中包含`-`属性的时候，插入时会自动过滤
		Name sql.NullString `json:"zhougunahgu"`
	}

	a1 := new(user)
	a1.Id = 1
	a1.Name = sql.NullString{"Mike", true}

	// r := db.NewQuery().Table("ttt").InsertSQL(a1)
	r := db.NewQuery().Table("ttt").Where("aaa", 12345678).Where("ccc", 123456789).UpdateSQL(a1)
	fmt.Println("result:", r)
	// fmt.Println("result:", e)
}
