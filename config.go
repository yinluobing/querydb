package querydb

import "time"

type DBConfig struct {
	Name         string //数据库连接别名
	IsMaster     bool   //是否是主库
	Driver       string
	Dsn          string
	MaxLifetime  time.Duration
	MaxIdleConns int
	MaxOpenConns int
}

type KConfig struct {
	TablePrefix  string
	StructTag    string
	DBConfigList []DBConfig
}
