# gosql 一个简单的MySql DML语句封装
---
### 列子
### 创建连接,连接池配置,设置表名称和数据写入
```Go
package main

import (
    "github.com/hfdend/gosql"
    "fmt"
)

func main() {
    // 建立连接
    m := gosql.NewDbMysql("127.0.0.1", 3306, "root", "", "test")
    // 设置最大连接数
    m.SetMaxOpenConns(30)
    // SetMaxIdleConns sets the maximum number of connections in the idle
    m.SetMaxIdleConns(10)
    // 设置空闲连接池的生存时间
    m.SetAutoCloseTime(100)
    // 设置表名
    m.SetTableName("user")

    // 数据插入
    data := map[string]interface{} {
        "user": "张三",
        "sex": "1",
        "age": 56,
        "hobbies": "乒乓球",
    }
    id, e := m.Insert(data);
    fmt.Println(e)
    fmt.Println(id)
    fmt.Println(m.LastSql)
}
```

### 简单查询操作与查询结果数据转换
```Go
condition := m.NewCondition()
condition.SetFilter("Id", 3)
condition.SetFilter("sex", 1)
condition.SetFilterEx("age", ">", 20)

// sql: select * from `user` where `Id` = 3 and `sex` = 1 and `age` > 20
r, e := m.SetCondition(condition).FindAll()

fmt.Println("错误", e)
fmt.Println("查询结果", r)
// 将数据转换成map
fmt.Println("数据转换成map", r.Result())
// 将数据Scan到结构体
type User struct {
    User        string  `field:"user"`
    Sex         string   `field:"sex"`
    Age         string  `field:"age"`
    Hobbies     string  `field:"Hobbies"`
}
var userAry []User
var userAryPtr []*User
r.Scan(&userAry)
r.Scan(&userAryPtr)
fmt.Println("数据转换成结构体", userAry, userAryPtr)
// 转换单个数据到结构体
for _, v := range r.ResultValue() {
    var user User
    v.Scan(&user)
    fmt.Println("单个结构体转换", user)
}


fmt.Println("执行的sql", m.LastSql)
```

