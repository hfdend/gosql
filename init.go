package db
import (
	"time"
)

var (
    AutoCloseTime = 100
)

func init ()  {
	// 监听mysql连接, 释放长时间没有使用的连接
	go func() {
		for {
			if MysqlDbMap != nil {
				for dataSourceName, homeDb :=  range MysqlDbMap {
					diffTime := int(time.Now().Unix() - homeDb.LastUseTime.Unix())
					if diffTime > 30 {
						homeDb.Close()
                        homeDb.IsClose = true
						homeDb = nil
						delete(MysqlDbMap, dataSourceName)
					}
				}
			}
            time.Sleep(time.Second * 10)
		}
	}()
    go func() {
        if RedisMap != nil {
            for addr, homeRedis :=  range RedisMap {
                diffTime := int(time.Now().Unix() - homeRedis.LastUseTime.Unix())
                if diffTime > 100 {
                    homeRedis.Close()
                    homeRedis = nil
                    delete(RedisMap, addr)
                }
            }
        }
        time.Sleep(time.Second * 10)
    }()
}
