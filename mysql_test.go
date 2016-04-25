package db

import (
    "testing"
    "fmt"
    "github.com/go-sql-driver/mysql"
)

type Model struct {
    Aaa  int8 `field:"Id"`
    Url string
    Url1 string
    test string
    Name []string
    CompanyName string `field:"CompanyName"`
}

func Test_Mysql(t *testing.T)  {
    var models []Model
    m := NewDbMysql("127.0.0.1", 3306, "root", "", "test")
    m.SetTableName("Sichuan")
    v, e := m.Limit(10).Order("Id desc").FindAll()
    m.Where(map[string])
    fmt.Println(e)
    v.Scan(&models)
    fmt.Printf("%#v\n", m)
}
