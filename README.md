# gosql

### 一个简单的sql语句封装

### 列子
```Go
// 建立连接
m := NewDbMysql("127.0.0.1", 3306, "root", "", "test")
// 设置表名
m.SetTableName("tablename")
```