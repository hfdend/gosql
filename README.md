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
// 自定义结构转换
handlerUserAry := &[]*User{}
r.Scan(handlerUserAry, func(v interface{}, row gosql.Value) {
    u := v.(*User)
    u.Age = "1010"
})
fmt.Println("自定义转换结果 start")
for _, v := range *handlerUserAry {
    fmt.Println(v)
}
fmt.Println("自定义转换结果 end")
fmt.Println("执行的sql", m.LastSql)
```
```Shell
错误 <nil>
查询结果 &[0xc82002c038]
数据转换成map [map[Id:3 user:张三 sex:1 age:28 hobbies:乒乓球]]
数据转换成结构体 [{张三 1 28 }] [0xc820010480]
单个结构体转换 {张三 1 28 }
自定义转换结果 start
&{张三 1 1010 }
自定义转换结果 end
执行的sql &{select * from `user` where `Id` = ? and `sex` = ? and `age` > ? [3 1 20]}
```

### OR查询,连表查询与分页查询
```go
// OR查询条件设置
conditon1 := m.NewCondition()
conditon2 := m.NewCondition()
condition := m.NewCondition()
conditon1.SetFilter("id", 1)
conditon2.SetFilter("id", 2)
condition.SetFilterOr(conditon1, conditon2)
// sql: where id = 1 or id = 2
m.SetCondition(condition)

// 关联查询
m.LeftJoin("user2", "user2.user_id = user.user_id")

// 分页
pager := m.NewPager()
// 设置每页条数
pager.Limit = 20
// 如果打开将使用子查询查询出总数
pager.IsSubqueries = false
// 设置偏移量
pager.Offset = 5
// 实现分页查询
values, err := m.PagerFindAll(pager)
fmt.Println("错误", err)
fmt.Println("总条数与分页情况", pager)
fmt.Println("当前页数的数据", values)
```
