package db

import (
    "testing"
    "fmt"
    "log"
)

func Test_Mysql(t *testing.T)  {
    // 建立连接
    m := NewDbMysql("127.0.0.1", 3306, "root", "", "test")
    // 设置最大连接数
    m.SetMaxOpenConns(30)
    // SetMaxIdleConns sets the maximum number of connections in the idle
    m.SetMaxIdleConns(10)
    // 设置连接自动关闭时间,如果一个连接在100秒内没有任何操作将会被自动关闭掉; 注意:如果一个SQL在100秒内没有执行完毕也会被关闭掉
    m.SetAutoCloseTime(100)
    // 设置表面
    m.SetTableName("user")

    // 数据插入
    data := map[string]interface{} {
        "user": "张三",
        "sex": "1",
        "age": 56,
        "hobbies": "乒乓球",
    }
    id, e := m.Insert(data);
    log.Println(e)
    fmt.Println(id)
}

func Test_Find(t *testing.T) {
    // 建立连接
    m := NewDbMysql("127.0.0.1", 3306, "root", "", "test")
    // 设置最大连接数
    m.SetMaxOpenConns(30)
    // SetMaxIdleConns sets the maximum number of connections in the idle
    m.SetMaxIdleConns(10)
    // 设置连接自动关闭时间,如果一个连接在100秒内没有任何操作将会被自动关闭掉; 注意:如果一个SQL在100秒内没有执行完毕也会被关闭掉
    m.SetAutoCloseTime(100)
    // 设置表面
    m.SetTableName("user")

    condition := m.NewCondition()
    c1 := m.NewCondition().SetFilter("Id", 3)
    c2 := m.NewCondition().SetFilter("Id", 4)
    c2.SetFilterEx("Age", ">", 20)

    condition.SetFilterOr(c1, c2)
    r, e := m.SetCondition(condition).FindAll()
    fmt.Println(e)
    fmt.Println(r)
    fmt.Println(m.LastSql)
}
