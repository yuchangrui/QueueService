package network

import (
	"QueueService/pkg/sortedset"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	addr string
	port int

	capSize int64                // 系统容量
	UserMap sync.Map             // 用户信息 userid -> Session
	queue   *sortedset.SortedSet // 队列存储
)

type Server struct {
	listener   net.Listener  // 监听句柄
	isStop     int32         // 服务是否已关闭, >0:已关闭
	serverExit chan struct{} // 服务退出chan
}

func (s *Server) Init() {
	s.isStop = 0
	s.serverExit = make(chan struct{})
	parseArgs() // 解析参数
	initQueue() // 初始化db
}

func (s *Server) Start() {
	// 启动listen
	var err error
	addrStr := fmt.Sprintf("%v:%v", addr, port)
	if s.listener, err = net.Listen("tcp", addrStr); err != nil {
		fmt.Printf("Listening error,  err:%v\n", err)
		return
	}
	log.Printf("Listening on %s.\n", addrStr)
	go s.startListen()

	//pprof
	go func() {
		if err := http.ListenAndServe(":13000", nil); err != nil {
			log.Printf("pprof init error, err:%v", err)
		}
	}()

	// 信号监听
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	// 定时打印服务状态
	watcher := time.NewTicker(time.Second) // 监听服务状态
	defer func() {
		s.stop()
		watcher.Stop()
	}()

	for {
		select {
		case <-watcher.C:
			{
				// 打印监听统计数据
				log.Printf("当前系统排队用户:%v\n", queue.Len())
			}
		case sig := <-ch:
			{
				log.Printf("捕获信号:%v, 服务即将退出...\n", sig)
				return
			}
		case <-s.serverExit:
			{
				return
			}
		}
	}
}

func (s *Server) stop() {
	if ok := atomic.CompareAndSwapInt32(&s.isStop, 0, 1); ok {
		close(s.serverExit)
		exitClear()

		log.Printf("server exit in 3 secs.\n")
		time.Sleep(time.Second * 3)
		log.Printf("client all exit, Goroutines:%v\n", runtime.NumGoroutine())
	}
}

func (s *Server) startListen() {
	defer s.stop()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				log.Printf("Accept Temporary, err: %v\n", err)
				continue
			} else {
				log.Printf("Accept error, err: %v\n", err)
				return
			}
		}

		s := NewSession(conn)
		go s.Start()
	}
}
