package gosql

import (
    "fmt"
    "log"
    "os"
    "testing"
    "time"
)

var MysqlIP = "127.0.0.1"
var MysqlPort = 3306
var DbUser = "root"
var DbPassword = "root"
var DbName = "test"

func testFindAndCountLimit(t *testing.T, db *DbMysql) {
    condition := db.NewCondition()
    condition.SetFilter("id", 3)
    condition.SetFilter("sex", 1)
    ret, err := db.SetCondition(condition).From("t_test").FindAll()
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(ret.Result())
    s_ret, err := db.SetCondition(condition).From("t_test").FindOne()
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(s_ret.Result())
    count, err := db.SetCondition(condition).From("t_test").Count()
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(count)
    ret, err = db.SetCondition(condition).From("t_test").Limit(0, 5).FindAll()
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(ret.Result())
}

func testGroupOrderHaving(t *testing.T, db *DbMysql) {
    condition := db.NewCondition()
    condition.SetFilterEx("age", ">", 3)
    ret, err := db.From("t_test").SetCondition(condition).Group("sex").FindAll()
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(ret.Result())
    ret, err = db.From("t_test").Order("age desc").FindAll()
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(ret.Result())

    var having_cond = map[string]interface{}{
        "count(sex) > ": 1,
    }
    ret, err = db.From("t_test").Group("sex").Having(having_cond).FindAll()
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(ret.Result())
}

func testDelExecQueryInsert(t *testing.T, db *DbMysql) {
    condition := db.NewCondition()
    condition.SetFilter("age", 13)
    del_count, err := db.SetCondition(condition).From("t_test").Del()
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(del_count)

    exec_count, err := db.Exec("update t_test set name = \"vicky_cheng\" where age = 12;")
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(exec_count)

    if r, e := db.Query("select * from t_test"); e != nil {
        t.Fatal(err)
    } else if r != nil {
        fmt.Println("result", r.Result())
    }

    // 数据插入
    data := map[string]interface{}{
        "name": "vicky",
        "sex":  "1",
        "age":  57,
    }
    id, err := db.Insert(data)
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(id)

    // 数据插入
    data = map[string]interface{}{
        "name": "vicky",
        "sex":  "1",
        "age":  57,
    }
    id, err = db.InsertIgnore(data)
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(id)
}

func testLockTable(t *testing.T, db *DbMysql) {
    err := db.LockReadTable("t_test")
    if err != nil {
        t.Fatal(err)
    }

    err = db.UnLockTable()
    if err != nil {
        t.Fatal(err)
    }
}

func testUpdate(t *testing.T, db *DbMysql) {
    condition := db.NewCondition()
    condition.SetFilter("id", 3)
    condition.SetFilter("sex", 1)

    var update_val = map[string]interface{}{
        "name": "vicky_123",
    }
    u_num, err := db.From("t_test").SetCondition(condition).Update(update_val)
    if err != nil {
        t.Fatal(err)
    }
    fmt.Println(u_num)
}

func TestAllApi(t *testing.T) {
    LogSetOutput(os.Stdout)
    tag_ins := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
    tag_ins.SetTableName("t_test")
    tag_ins.Master.MaxOpenConns = 20
    tag_ins.Master.MaxIdleConns = 10
    testFindAndCountLimit(t, tag_ins)
    testGroupOrderHaving(t, tag_ins)
    testDelExecQueryInsert(t, tag_ins)
    testUpdate(t, tag_ins)
    testLockTable(t, tag_ins)
    os.Exit(0)
}

func Test_Mysql(t *testing.T) {
    // 建立连接
    tag_ins := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
    tag_ins.SetTableName("t_test")
    // 设置最大连接数
    // SetMaxIdleConns sets the maximum number of connections in the idle
    // 设置连接自动关闭时间,如果一个连接在100秒内没有任何操作将会被自动关闭掉; 注意:如果一个SQL在100秒内没有执行完毕也会被关闭掉
    // 设置表面
    tag_ins.SetTableName("t_test")

    // 数据插入
    data := map[string]interface{}{
        "name": "张三",
        "sex":  "1",
        "age":  56,
    }
    id, e := tag_ins.Insert(data)
    log.Println(e)
    fmt.Println(id)
}

func Test_Connect(t *testing.T) {
    m := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
    m.SetTableName("t_test")
    for {
        _, e := m.Exec("select * from user0")
        log.Println(e)
    }
    time.Sleep(time.Hour)
}

func Benchmark_Connect(b *testing.B) {
    for i := 0; i < b.N; i++ {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        if r, e := db.Query("select * from t_test limit 1"); e != nil {
            fmt.Println("error", e)
        } else if r != nil {
            v := r.Result()
            fmt.Println("result", v)
        }
    }
}

func Benchmark_Mysql(b *testing.B) {
    c := make(chan int)
    t := 100

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            _, e := db.Insert(data)
            log.Println(e)

            c1 <- i
        }
    }(c)

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            _, e := db.Insert(data)
            log.Println(e)

            c1 <- i
        }
    }(c)

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            _, e := db.Insert(data)
            log.Println(e)

            c1 <- i
        }
    }(c)

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            _, e := db.Insert(data)
            log.Println(e)

            c1 <- i
        }
    }(c)

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            _, e := db.Insert(data)
            log.Println(e)

            c1 <- i
        }
    }(c)

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            _, e := db.Insert(data)
            log.Println(e)

            c1 <- i
        }
    }(c)

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            db.Insert(data)

            c1 <- i
        }
    }(c)

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            _, e := db.Insert(data)
            log.Println(e)

            c1 <- i
        }
    }(c)

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            _, e := db.Insert(data)
            log.Println(e)

            c1 <- i
        }
    }(c)

    go func(c1 chan int) {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < t; i++ {
            // 数据插入
            data := map[string]interface{}{
                "name": fmt.Sprintf("tony%d", i),
                "sex":  1,
                "age":  56,
            }
            _, e := db.Insert(data)
            log.Println(e)

            c1 <- i
        }
    }(c)

    for i := 0; i < t*10; i++ {
        <-c
    }
}

func Benchmark_Query(b *testing.B) {
    c := make(chan int, 3)
    go func() {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < b.N; i++ {
            condition := db.NewCondition()
            condition.SetFilter("id", 3)
            condition.SetFilter("sex", 1)
            ret, err := db.SetCondition(condition).From("t_test").FindAll()
            if err != nil {
                log.Println(err)
            } else {
                fmt.Println(ret.Result())
            }

        }
        c <- 1
    }()
    go func() {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < b.N; i++ {
            condition := db.NewCondition()
            condition.SetFilter("id", 3)
            condition.SetFilter("sex", 1)
            ret, err := db.SetCondition(condition).From("t_test").FindAll()
            if err != nil {
                log.Println(err)
            } else {
                fmt.Println(ret.Result())
            }
        }
        c <- 2
    }()
    go func() {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < b.N; i++ {
            condition := db.NewCondition()
            condition.SetFilter("id", 3)
            condition.SetFilter("sex", 1)
            ret, err := db.SetCondition(condition).From("t_test").FindAll()
            if err != nil {
                log.Println(err)
            } else {
                fmt.Println(ret.Result())
            }
        }
        c <- 3
    }()

    for i := 0; i < 3; i++ {
        <-c
    }
}

func Benchmark_Update(b *testing.B) {
    c := make(chan int, 3)
    go func() {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < b.N; i++ {
            condition := db.NewCondition()
            condition.SetFilter("id", 3)
            condition.SetFilter("sex", 1)

            var update_val = map[string]interface{}{
                "name": "vicky_123",
            }
            num, err := db.From("t_test").SetCondition(condition).Update(update_val)
            if err != nil {
                log.Println(err)
            }
            fmt.Printf("Update rows:%d\n", num)
        }
        c <- 1
    }()

    go func() {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < b.N; i++ {
            condition := db.NewCondition()
            condition.SetFilter("id", 3)
            condition.SetFilter("sex", 1)

            var update_val = map[string]interface{}{
                "name": "vicky_123",
            }
            num, err := db.From("t_test").SetCondition(condition).Update(update_val)
            if err != nil {
                log.Println(err)
            }
            fmt.Printf("Update rows:%d\n", num)
        }
        c <- 2
    }()

    go func() {
        db := NewDbMysql(MysqlIP, MysqlPort, DbUser, DbPassword, DbName)
        db.SetTableName("t_test")
        for i := 0; i < b.N; i++ {
            condition := db.NewCondition()
            condition.SetFilter("id", 3)
            condition.SetFilter("sex", 1)

            var update_val = map[string]interface{}{
                "name": "vicky_123",
            }
            num, err := db.From("t_test").SetCondition(condition).Update(update_val)
            if err != nil {
                fmt.Println(err)
            }
            fmt.Printf("Update rows:%d\n", num)
        }
        c <- 3
    }()

    for i := 0; i < 3; i++ {
        <-c
    }
}
