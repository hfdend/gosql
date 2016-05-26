package gosql

import (
    "testing"
    "fmt"
    "log"
    "time"
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

type User struct {
    Id      string      `field:"Id"`
    User    string      `field:"user"`
    Sex     string      `field:"sex"`
    Age     string      `field:"age"`
    Hobbies string      `field:"hobbies"`
    Price   uint     `field:"price"`
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

    for i := 0; i < 1; i++ {
        r, e := m.SetCondition(nil).FindAll()
        fmt.Println(e)

        var data []*User
        r.Scan(&data)
        for _, v := range data {
            fmt.Println(v)
        }
        fmt.Println(m.LastSql)

        time.Sleep(10 * time.Second)
    }
}

func Test_Tostring(t *testing.T) {
    type d struct {
        haha int
    }
    type M struct {
        D   *d
        Id      int8         `field:"id"`
        Name    string      `field:"name"`
        Name2    float64      `field:"name2"`
    }

    m := &M{D:&d{12}, Id:13, Name:"哈哈", Name2: 123.433}
    s := toString(m)
    fmt.Println(s)
}
