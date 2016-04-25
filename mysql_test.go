package db

import (
    "testing"
)

func Test_Mysql(t *testing.T)  {
    // 建立连接
    m := NewDbMysql("127.0.0.1", 3306, "root", "", "test")
    // 设置最大连接数
    m.DB.SetMaxOpenConns(30)
    // SetMaxIdleConns sets the maximum number of connections in the idle
    m.DB.SetMaxIdleConns(10)
    // 设置连接自动关闭时间,如果一个连接在100秒内没有任何操作将会被自动关闭掉; 注意:如果一个SQL在100秒内没有执行完毕也会被关闭掉
    m.SetAutoCloseTime(100)
    // 设置表面
    m.SetTableName("user")
}
