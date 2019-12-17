## 简介

querydb是一个源于Laravel的ORM框架

## 用法示例

```go
package main

import (
	"fmt"
	"github.com/pm-esd/querydb"
)

func main() {
    //初始化配置
    kConf := new(querydb.KConfig)
    //初始化DB的配置
    dbConfig := new(querydb.DBConfig)
    dbConfig.Driver = "mysql"
    dbConfig.Dsn = "root:123456@tcp(127.0.0.1:3306)/querydb?charset=utf8&parseTime=true"
    dbConfig.IsMaster = true
    kConf.DBConfigList = []querydb.DBConfig{*dbConfig}
    querydb.RegisterDataBase(*kConf)

    //原生SQL查询
    querydb.Select("select * from user where id = ?", 1).ToArray()
    //返回map[string][string]
    querydb.Select("select * from user where id = ?", 1).ToMap()
    //返回struct
    type user struct {
        Id int `db:"id;auto"`
        Name string `db:"name"`
    }
    var result []user
    querydb.Select("select * from user").ToStruct(&result)
    fmt.Println("result:", result)

    //链式操作,返回单条数据
    querydb.Table("user").Where("id", 1).First().ToArray()

    //支持指定库操作
    var u user
    querydb.WithDB("mysql::master").Table("user").Where("id", 1).First().ToStruct(&u)

    //批量插入支持map方式和struct方式
    a1 := new(user)
    a1.Name = "张三"

    a2 := new(user)
    a2.Name = "李四"

    users := []user{*a1, *a2}
    querydb.Table("user").MultiInsert(users)
}
```

### 查询数据
```go

//查询单条数据
//返回[]string
arr, err := querydb.Table("user").Where("id", 1).First().ToArray()

//返回map[string][string
mp, err := querydb.Table("user").Where("id", 1).First().ToMap()

type user struct {
    Id int `db:"id"`
    Name string `db:"name"`
}

//返回结构体
var result user
err := querydb.Table("user").Where("id", 1).First().ToStruct(&result)


//查询多条数据
//返回[][]string
arr, err := querydb.Table("user").Where("id", 1).Get().ToArray()

//返回[]map[string][string
mp, err := querydb.Table("user").Where("id", 1).Get().ToMap()

type user struct {
    Id int `db:"id"`
    Name string `db:"name"`
}

//返回结构体
var result []user
err := querydb.Table("user").Where("id", 1).Get().ToStruct(&result)

```

### 插入数据
```go

//通过结构体插入
type user struct {
	Id int `db:"id;auto"`   //tag中包含auto属性的时候，插入时会自动过滤
	Name string `db:"name"`
}

a1 := new(user)
a1.Name = "张三"

//插入单条
querydb.Table("user").Insert(a1)

//插入多条
a1 := new(user)
a1.Name = "张三"

a2 := new(user)
a2.Name = "李四"
users := []user{*a1, *a2}
querydb.Table("user").MultiInsert(users)

//通过map方式插入
user := make(map[string]string)
user["name"] = "张三"
querydb.Table("user").Insert(user)

```


### 更新数据
```go

data := make(map[string]interface{})
data["name"] = "李四"

querydb.Table("user").Where("id", 1).Update(data)

```

### 删除数据
```go
querydb.Table("user").Where("id", 1).Delete()
```
