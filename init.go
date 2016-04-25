package db
import (
	"time"
)

func init ()  {
	// 监听mysql连接, 释放长时间没有使用的连接
	go func() {
		for {
			if MysqlDbMap != nil {
				for dataSourceName, homeDb :=  range MysqlDbMap {
					diffTime := int(time.Now().Unix() - homeDb.LastUseTime.Unix())
					if homeDb.AutoCloseTime > 0 && diffTime > homeDb.AutoCloseTime {
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
}
