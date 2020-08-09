package main

import (
	"flag"
	"log"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	wg sync.WaitGroup

	addr           string
	port           int
	num            int
	customer       bool   // 是否为消费者
	consumSpeed    int64  // 消费速率
	priority       bool   // 是否为管理员调整用户优先级
	prioritySecs   int64  // 调整用户优先级秒数
	priorityUserId string // 调整优先级用户ID
)

func main() {
	parseArgs() // 解析参数

	// 启动 消费 or 生产 or 管理员
	if customer {
		startCustom()
	} else if priority {
		changUserpriority()
	} else {
		startProduct()
	}

	wg.Wait()
	time.Sleep(time.Second)
	log.Printf("client all exit, Goroutines:%v\n", runtime.NumGoroutine())
}

func parseArgs() {
	flag.StringVar(&addr, "addr", "localhost", "启动地址,默认为localhost.")
	flag.IntVar(&port, "port", 8000, "启动端口,默认为8000.")
	flag.IntVar(&num, "num", 10, "启动排队客户端数量,默认为10.")
	flag.BoolVar(&customer, "customer", false, "是否为消费端,默认为false.")
	flag.Int64Var(&consumSpeed, "consumSpeed", 10, "消费端消费速率,默认为10.")
	flag.BoolVar(&priority, "priority", false, "是否为管理员调整用户优先级,默认为false.")
	flag.Int64Var(&prioritySecs, "prioritySecs", 10, "调整用户优先级秒数,默认为10.")
	flag.StringVar(&priorityUserId, "priorityUserId", "", "调整优先级用户ID,默认为localhost.")
	flag.Parse()
	log.Printf("client start, run args:\nhost:%v\nport:%v\nnum:%v\ncustomer:%v\nconsumSpeed:%v\npriority:%v\nprioritySecs:%v\npriorityUserId:%v\n", addr, port, num, customer, consumSpeed, priority, prioritySecs, priorityUserId)
}

// 生产者
func startProduct() {
	dest := addr + ":" + strconv.Itoa(port)
	for i := 0; i < num; i++ {
		wg.Add(1)
		time.Sleep(time.Millisecond * time.Duration(randN(1, 10))) // 延迟,防治对服务冲击
		go func() {
			conn, err := net.DialTimeout("tcp", dest, time.Second*5)
			if err != nil {
				if _, t := err.(*net.OpError); t {
					log.Printf("Some problem connecting.err:%v\n", err)
				} else {
					log.Printf("Unknown error: err:%v\n" + err.Error())
				}
				wg.Done()
				return
			}
			cli := newClient(conn, 1)
			go cli.start()
		}()
	}
}

// 消费者
func startCustom() {
	dest := addr + ":" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", dest, time.Second*2)
	if err != nil {
		if _, t := err.(*net.OpError); t {
			log.Printf("Some problem connecting.err:%v\n", err)
		} else {
			log.Printf("Unknown error: \n" + err.Error())
		}
		return
	}
	wg.Add(1)
	cli := newClient(conn, 2)
	go cli.start()
}

func changUserpriority() {
	dest := addr + ":" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", dest, time.Second*2)
	if err != nil {
		if _, t := err.(*net.OpError); t {
			log.Printf("Some problem connecting.err:%v\n", err)
		} else {
			log.Printf("Unknown error: \n" + err.Error())
		}
		return
	}
	wg.Add(1)
	cli := newClient(conn, 3)
	go cli.start()
}
