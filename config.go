package querydb

import (
	"context"
	"database/sql"
	"math/rand"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

//Config 数据库配置
type Config struct {
	Username     string        //账号 root
	Password     string        //密码
	Host         string        //host localhost
	Port         string        //端口 3306
	Charset      string        //字符编码 utf8mb4
	Database     string        //默认连接数据库
	MaxLifetime  time.Duration //设置一个连接的最长生命周期，因为数据库本身对连接有一个超时时间的设置，如果超时时间到了数据库会单方面断掉连接，此时再用连接池内的连接进行访问就会出错, 因此这个值往往要小于数据库本身的连接超时时间
	MaxIdleConns int           //设置闲置的连接数,连接池里面允许Idel的最大连接数, 这些Idel的连接 就是并发时可以同时获取的连接,也是用完后放回池里面的互用的连接, 从而提升性能
	MaxOpenConns int           //设置最大打开的连接数，默认值为0表示不限制。控制应用于数据库建立连接的数量，避免过多连接压垮数据库。
	Slave        []*Config     //从库
}

//SetSlave 设置 Slave
func (config *Config) SetSlave(c *Config) *Config {
	if config.Slave == nil {
		config.Slave = make([]*Config, 0)
	}
	config.Slave = append(config.Slave, c)
	return config
}

// Configs 配置
type Configs struct {
	cfg         map[string]*Config
	connections map[string]*QueryDb
}

//Default ..
func Default() *Configs {
	return &Configs{
		cfg:         make(map[string]*Config),
		connections: make(map[string]*QueryDb),
	}
}

//SetConfig 设置配置文件
func (configs *Configs) SetConfig(name string, cf *Config) *Configs {
	configs.cfg[name] = cf
	return configs
}

//URI 构造数据库连接
func (config *Config) URI() string {
	return config.Username + ":" +
		config.Password + "@tcp(" +
		config.Host + ":" +
		config.Port + ")/" +
		config.Database + "?charset=" +
		config.Charset + "&loc=" + time.Local.String()
}

//random 随机数
func random(max int) int {
	if max < 1 {
		return 0
	}
	rand.Seed(time.Now().Unix())
	return rand.Intn(max)
}

//Write 获取一个
func (configs *Configs) Write(name string) *QueryDb {
	config, ok := configs.cfg[name]
	if !ok {
		logrus.Fatal("DB配置:" + name + "找不到！")
	}
	//获取主
	db := connect(config)
	configs.connections[name] = &QueryDb{db: db, ctx: context.Background()}
	return configs.connections[name]
}

func (configs *Configs) Read(name string) *QueryDb {
	config, ok := configs.cfg[name]
	if !ok {
		logrus.Fatal("DB配置:" + name + "找不到！")
	}
	keyname := name
	readlen := len(config.Slave)

	readnum := 0
	if readlen > 0 {
		readnum = random(readlen)
		keyname += "_read_" + strconv.Itoa(readnum)
		config = configs.cfg[name].Slave[readnum]
	}
	//获取主
	db := connect(config)
	configs.connections[keyname] = &QueryDb{db: db, ctx: context.Background()}
	return configs.connections[keyname]
}

//connect 数据库连接
func connect(config *Config) *sql.DB {
	//数据库连接
	db, err := sql.Open("mysql", config.URI())
	if err != nil {
		logrus.Fatal("DB连接错误！")
	}
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxLifetime > 0 {
		db.SetConnMaxLifetime(config.MaxLifetime)
	}
	return db
}
