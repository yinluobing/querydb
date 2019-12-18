# querydb

这是一个针对 go mysql 查询的查询构造器，支持主从配置，支持读写分离。

配置范例：

```go

//配置集合
    master := &querydb.Config{
        Username:        "root",
        Password:        "mysql",
        Host:            "127.0.0.1",
        Port:            "33061",
        Charset:         "utf8mb4",
        Database:        "ott",
        ConnMaxLifetime: 120,
        MaxIdleConns:    200,
        MaxOpenConns:    800,
    }
    slave1 := &querydb.Config{
        Username:        "root",
        Password:        "mysql",
        Host:            "127.0.0.1",
        Port:            "33061",
        Charset:         "utf8mb4",
        Database:        "ott",
        ConnMaxLifetime: 120,
        MaxIdleConns:    200,
        MaxOpenConns:    800,
    }
    slave2 := &querydb.Config{
        Username:        "root",
        Password:        "mysql",
        Host:            "127.0.0.1",
        Port:            "33061",
        Charset:         "utf8mb4",
        Database:        "ott",
        ConnMaxLifetime: 120,
        MaxIdleConns:    200,
        MaxOpenConns:    800,
    }
    master.SetSlave(slave1)
    master.SetSlave(slave2)
    instance := querydb.Default()
    instance.SetConfig("test", master)
    db := instance.Write("test") //主库
    db := instance.Read("test") //从库

    type user struct {
        Id int `json:"json"`
        Name string `json:"name"`
    }
    var result []user
    db.Table("user").Rows().ToStruct(&result)
    fmt.Println("result:", result)
```



### 查询数据
```go

//查询单条数据
//返回[]string
arr, err := db.Table("user").Where("id", 1).Row().ToArray()


//返回map[string][string
mp, err := db.Table("user").Where("id", 1).Row().ToMap()


type user struct {
    Id int `json:"id"`
    Name string `json:"name"`
}

//返回结构体
var result user
err := db.Table("user").Where("id", 1).First().ToStruct(&result)




//查询多条数据
//返回[][]string
arr, err := db.Table("user").Where("id", 1).Get().ToArray()

//返回[]map[string][string
mp, err := db.Table("user").Where("id", 1).Get().ToMap()

type user struct {
    Id int `json:"id"`
    Name string `json:"name"`
}
//返回结构体
var result []user
err := db.Table("user").Where("id", 1).Get().ToStruct(&result)

```





### 插入数据
```go

//通过结构体插入
type user struct {
	Id int `json:"-"`   //tag中包含`-`属性的时候，插入时会自动过滤
	Name string `json:"name"`
}

a1 := new(user)
a1.Name = "张三"

//插入单条
db.Table("user").Insert(a1)

//插入多条
a1 := new(user)
a1.Name = "张三"

a2 := new(user)
a2.Name = "李四"
users := []user{*a1, *a2}
db.Table("user").MultiInsert(users)

//通过map方式插入
user := make(map[string]string)
user["name"] = "张三"
db.Table("user").Insert(user)

```




### 更新数据
```go

data := make(map[string]interface{})
data["name"] = "李四"

db.Table("user").Where("id", 1).Update(data)

```

### 删除数据
```go
db.Table("user").Where("id", 1).Delete()
```


### 事务
```go

开启事务需要用主库

db.Begin()

db.Table("user").Where("id", 1).Delete()

db.Commit()

db.RollBack()
```
