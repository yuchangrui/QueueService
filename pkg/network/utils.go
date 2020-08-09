package network

import (
	"QueueService/pkg/sortedset"
	"flag"
	"log"
)

func initQueue() {
	queue = sortedset.Make()
}

func exitClear() {
	queue.RemoveByRank(0, queue.Len()) // 清除队列
	UserMap.Range(
		func(key, value interface{}) bool {
			s, ok := value.(*Session)
			if ok {
				s.Stop() // 关闭连接
			}
			UserMap.Delete(key)
			return true
		})
}

func parseArgs() {
	flag.StringVar(&addr, "addr", "localhost", "启动地址,默认为localhost.")
	flag.IntVar(&port, "port", 8000, "启动端口,默认为8000.")
	flag.Int64Var(&capSize, "capSize", 200000, "系统容量,默认为200000.")
	flag.Parse()
	log.Printf("server start, run args:\naddr:%v\nport:%v\ncapSize:%v\n", addr, port, capSize)
}
