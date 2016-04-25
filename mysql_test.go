package db

import (
    "testing"
    "fmt"
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
    var m []Model
    mysql := NewDbMysql("127.0.0.1", 3306, "root", "", "test")
    mysql.SetTableName("Sichuan")
    v, e := mysql.Limit(10).Order("Id desc").FindAll()
    fmt.Println(e)
    v.Scan(&m)
    fmt.Printf("%#v\n", m)
}
