package gosql

import (
    "math/rand"
    "time"
)

type ConfigModel struct {
    Host          string
    Port          int
    User          string
    Password      string
    DBName        string
    AutoCloseTime int
    MaxOpenConns  int
    MaxIdleConns  int
}

type Config struct {
    Master *ConfigModel
    Slave  []*ConfigModel
}

func (this *Config) NewDbMysql() *DbMysql {
    m := new(DbMysql)
    m.Master = this.Master
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    m.Slave = this.Slave[r.Intn(len(this.Slave))]
    return m
}
