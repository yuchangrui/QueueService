package network

import (
	"QueueService/pkg/msg"
	"log"
	"net"
	"sync/atomic"
	"time"
)

type Session struct {
	conn        net.Conn
	UserId      string
	isClose     int32         // 连接是否关闭 >0:关闭
	isQuitQueue int32         // 是否已出队列 >0:关闭
	closeChan   chan struct{} // 连接出错关闭
	sendMsgChan chan []byte   // 发送数据chan
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		conn:        conn,
		isClose:     0,
		isQuitQueue: 0,
		closeChan:   make(chan struct{}),
		sendMsgChan: make(chan []byte, 1024),
	}
}

func (s *Session) Start() {
	go s.recvMsg()
	go s.sendMsg()
}

func (s *Session) Stop() {
	if ok := atomic.CompareAndSwapInt32(&s.isClose, 0, 1); ok {
		s.conn.Close()
		close(s.closeChan)
		close(s.sendMsgChan)
		s.quitQueque()
	}
}

func (s *Session) recvMsg() {
	defer s.Stop()
	end := 0    // 有效数据长度
	lenght := 0 // 读取到的数据集长度
	var err error
	buf := make([]byte, 10240) // 数据池
	for {
		if err := s.conn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
			// log.Printf("SetReadDeadline error, err:%v, uId:%v\n", err, s.UserId)
			break
		}
		if lenght, err = s.conn.Read(buf[end:]); err != nil {
			// log.Println(err)
			return
		}

		end = end + lenght // 更新数据长度
		for {
			info, n, err := msg.UnPackMsg(buf, end)
			if err != nil {
				return
			}
			if n == 0 {
				// 没有完整的消息
				break
			}

			// 删除处理完的数据
			copy(buf, buf[n:end])
			end -= n // 调整index

			// 协议处理
			if handlerFunc, ok := mapHandleFunc[info.MsgId]; ok {
				go handlerFunc(info.Body, s)
			} else {
				return // 不支持的协议,直接断开,防治被渗透攻击
			}
		}
	}
}

func (s *Session) sendMsg() {
	defer s.Stop()
	for {
		select {
		case msg, ok := <-s.sendMsgChan:
			if ok {
				_ = s.conn.SetWriteDeadline(time.Now().Add(time.Second * 2)) // 设置写超时
				if _, err := s.conn.Write(msg); err != nil {
					// log.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				break
			}
		case <-s.closeChan:
			return
		}
	}
}

func (s *Session) sendRankResp() {
	rank, ok := s.getRank()
	if !ok {
		s.Stop()
		return
	}
	info, ok := msg.MakeRankAckMsg(s.UserId, rank)
	if !ok {
		return
	}
	if ic := atomic.LoadInt32(&s.isClose); ic > 0 {
		return
	}

	s.sendMsgChan <- info
	// log.Printf("sendRankResp , userId:%v, size:%v\n", s.UserId, s.getRank())
}

func (s *Session) sendFullResp() {
	info, ok := msg.MakeFullMsg()
	if !ok {
		return
	}
	if ic := atomic.LoadInt32(&s.isClose); ic > 0 {
		return
	}

	s.sendMsgChan <- info
	log.Printf("sendFullResp , userId:%v\n", s.UserId)
}

func (s *Session) sendMsgPopAckResp(size int64) {
	info, ok := msg.MakePopAckMsg(size)
	if !ok {
		return
	}

	if ic := atomic.LoadInt32(&s.isClose); ic > 0 {
		return
	}
	s.sendMsgChan <- info
	log.Printf("custom user size:%v\n", size)
}

func (s *Session) sendMsgChangRankAckResp(userId string, old, new int64) {
	info, ok := msg.MakeChangeRankAckMsg(userId, old, new)
	if !ok {
		return
	}

	if ic := atomic.LoadInt32(&s.isClose); ic > 0 {
		return
	}
	s.sendMsgChan <- info
	log.Printf("sendMsgChangRankAckResp , userId:%v, old:%v, new:%v\n", s.UserId, old, new)
}

func (s *Session) sendOutMsg() {
	info, ok := msg.MakeOutMsg()
	if !ok {
		log.Printf("MakeOutMsg error")
		return
	}
	if ic := atomic.LoadInt32(&s.isClose); ic > 0 {
		return
	}
	s.sendMsgChan <- info
	log.Printf("sendOutMsg , userId:%v\n", s.UserId)
}
