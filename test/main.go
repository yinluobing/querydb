package main

import (
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

	r, e := db.Table("ott_video").Rows().ToMap()
	fmt.Println("result:", r)
	fmt.Println("result:", e)
}
