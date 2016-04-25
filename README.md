# gosql 一个简单的sql语句封装
---
### 列子
### 创建连接,连接池配置与设置表名称
```Go
// 建立连接
m := NewDbMysql("127.0.0.1", 3306, "root", "", "test")
// 设置最大连接数
m.DB.SetMaxOpenConns(30)
// SetMaxIdleConns sets the maximum number of connections in the idle
m.DB.SetMaxIdleConns(10)
// 设置连接自动关闭时间,如果一个连接在100秒内没有任何操作将会被自动关闭掉; 注意:如果一个SQL在100秒内没有执行完毕也会被关闭掉
m.SetAutoCloseTime(100)
// 设置表名
m.SetTableName("user")
```

### 数据插入
```Go
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
```

### 查询
```Go
condition := m.NewCondition()
condition.SetFilter("Id", 3)
condition.SetFilter("sex", 1)
condition.SetFilterEx("age", ">", 20)

// sql: select * from `user` where `Id` = 3 and `sex` = 1 and `age` > 20
r, e := m.SetCondition(condition).FindAll()

fmt.Println(e)
fmt.Println(r)
fmt.Println(m.LastSql)
```
