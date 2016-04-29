package gosql
import (
    "time"
    "sync"
)

func init ()  {
    // 监听mysql连接, 释放长时间没有使用的连接
    var mutex sync.Mutex
    go func() {
        for {
            if MysqlDbMap != nil {
                for dataSourceName, homeDb :=  range MysqlDbMap {
                    go func(dataSourceName string, homeDb *homeDB) {
                        diffTime := int(time.Now().Unix() - homeDb.LastUseTime.Unix())
                        if homeDb.AutoCloseTime > 0 && diffTime >= homeDb.AutoCloseTime {
                            homeDb.Close()
                            homeDb.IsClose = true
                            homeDb = nil
                            mutex.Lock()
                            defer mutex.Unlock()
                            delete(MysqlDbMap, dataSourceName)
                        }
                    }(dataSourceName, homeDb)
                }
            }
            time.Sleep(time.Second * 10)
        }
    }()
}
