package main

import (
	"QueueService/pkg/msg"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	uuid "github.com/satori/go.uuid"
)

var (
	u1 uuid.UUID
)

func init() {
	u1 = uuid.Must(uuid.NewV4(), nil)
}

type Clint struct {
	Conn           net.Conn
	UserId         string
	flag           int    // 1:生产者, 2:消费者, 3:管理员
	ConsumSize     int64  // 消费速率, 秒
	prioritySecs   int64  // 调整用户优先级秒数
	priorityUserId string // 调整优先级用户ID

	isClosed    int32         // 是否已关闭, >0:已关闭
	closeChan   chan struct{} // 关闭连接
	sendMsgChan chan []byte   // 发送数据
}

// randN 取随机数,用于时间延迟
func randN(start, end int) int {
	return start + rand.Intn(end-start+1)
}

func newClient(conn net.Conn, flag int) *Clint {
	u1 := uuid.Must(uuid.NewV4(), nil)
	ret := &Clint{
		Conn:           conn,
		UserId:         u1.String(),
		flag:           flag,
		ConsumSize:     consumSpeed,
		closeChan:      make(chan struct{}),
		isClosed:       0,
		sendMsgChan:    make(chan []byte, 1024),
		prioritySecs:   prioritySecs,
		priorityUserId: priorityUserId,
	}
	return ret
}

func (c *Clint) start() {
	if c.flag == 1 {
		c.startProduct()
	} else if c.flag == 2 {
		c.startConsum()
	} else {
		c.administrator()
	}
}

func (c *Clint) Close() {
	if ok := atomic.CompareAndSwapInt32(&c.isClosed, 0, 1); ok {
		_ = c.Conn.Close()
		close(c.closeChan)
		close(c.sendMsgChan)
	}
}

func (c *Clint) startProduct() {
	defer func() {
		c.Close()
		wg.Done()
	}()

	time.Sleep(time.Millisecond * time.Duration(randN(100, 3000)))
	if err := c.sendJoinMsg(); err != nil {
		log.Printf("clint.start sendJoinMsg error:%v\n", err)
		return
	}

	go c.recvMsg()
	go c.sendMsg()

	// 启动定时器
	ticker := time.NewTicker(time.Millisecond * time.Duration(randN(1000, 2000)))
	rankTicker := time.NewTicker(time.Millisecond * time.Duration(randN(1000, 2000)))

	defer func() {
		ticker.Stop()
		rankTicker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if err := c.sendHeartBeatMsg(); err != nil {
				log.Printf("sendHeartBeatMsg error:%v\n", err)
				return
			}
		case <-rankTicker.C:
			if err := c.sendRankMsg(); err != nil {
				log.Printf("sendRankMsg error:%v\n", err)
				return
			}
		case <-c.closeChan:
			{
				return
			}
		}
	}
}

func (c *Clint) startConsum() {
	defer func() {
		c.Close()
		wg.Done()
	}()

	go c.recvMsg()
	go c.sendMsg()

	// 启动定时器
	ticker := time.NewTicker(time.Millisecond * time.Duration(randN(1000, 2000)))
	cusTicker := time.NewTicker(time.Millisecond * time.Duration(randN(1000, 2000)))

	defer func() {
		ticker.Stop()
		cusTicker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if err := c.sendHeartBeatMsg(); err != nil {
				log.Printf("sendHeartBeatMsg error:%v\n", err)
				return
			}
		case <-cusTicker.C:
			if err := c.sendQueuePopMsg(c.ConsumSize); err != nil {
				log.Printf("sendHeartBeatMsg error:%v\n", err)
				return
			}
		case <-c.closeChan:
			{
				return
			}
		}
	}
}

func (c *Clint) administrator() {
	defer func() {
		wg.Done()
	}()

	c.UserId = c.priorityUserId
	if err := c.sendChangeRankMsg(); err != nil {
		log.Printf("clint.start sendRankMsg error:%v\n", err)
		return
	}

	go c.recvMsg()
	go c.sendMsg()

	// 启动定时器
	ticker := time.NewTicker(time.Millisecond * time.Duration(randN(1000, 2000)))
	rankTicker := time.NewTicker(time.Millisecond * time.Duration(randN(1000, 2000)))

	defer func() {
		ticker.Stop()
		rankTicker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if err := c.sendHeartBeatMsg(); err != nil {
				log.Printf("sendHeartBeatMsg error:%v\n", err)
				return
			}
		case <-rankTicker.C:
			if err := c.sendRankMsg(); err != nil {
				log.Printf("sendRankMsg error:%v\n", err)
				return
			}
		case <-c.closeChan:
			{
				return
			}
		}
	}
}

func (c *Clint) recvMsg() {
	defer c.Close()
	end := 0    // 有效数据长度
	lenght := 0 // 读取到的数据集长度
	var err error
	buf := make([]byte, 10240) // 数据池
	for {
		if err := c.Conn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
			// log.Printf("SetReadDeadline error, err:%v, uId:%v\n", err, c.UserId)
			return
		}
		if lenght, err = c.Conn.Read(buf[end:]); err != nil {
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
				go handlerFunc(info.Body, c)
			} else {
				return // 不支持的协议,直接断开,防治被渗透攻击
			}
		}

	}
}

func (c *Clint) sendMsg() {
	defer c.Close()
	for {
		select {
		case msg, ok := <-c.sendMsgChan:
			if ok {
				_ = c.Conn.SetWriteDeadline(time.Now().Add(time.Second * 2)) // 设置写超时
				if _, err := c.Conn.Write(msg); err != nil {
					// log.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				// log.Println("msgBuffChan is Closed")
				return
			}
		case <-c.closeChan:
			return
		}
	}
}

func (c *Clint) sendJoinMsg() error {
	info, ok := msg.MakeJoinMsg(c.UserId)
	if !ok {
		return fmt.Errorf("sendJoinMsg.MakeJoinMsg, uId:%v\n", c.UserId)
	}
	if ic := atomic.LoadInt32(&c.isClosed); ic > 0 {
		return nil
	}
	c.sendMsgChan <- info
	return nil
}

func (c *Clint) sendHeartBeatMsg() error {
	info, ok := msg.MakeHeartBeatMsg()
	if !ok {
		return fmt.Errorf("sendJoinMsg.MakeJoinMsg, uId:%v\n", c.UserId)
	}

	if ic := atomic.LoadInt32(&c.isClosed); ic > 0 {
		return nil
	}
	c.sendMsgChan <- info
	return nil
}

func (c *Clint) sendRankMsg() error {
	info, ok := msg.MakeRankMsg(c.UserId)
	if !ok {
		return fmt.Errorf("sendJoinMsg.MakeJoinMsg, uId:%v\n", c.UserId)
	}
	if ic := atomic.LoadInt32(&c.isClosed); ic > 0 {
		return nil
	}
	c.sendMsgChan <- info
	return nil
}

func (c *Clint) sendQueuePopMsg(size int64) error {
	info, ok := msg.MakePopMsg(size)
	if !ok {
		return fmt.Errorf("MakePopMsg error, msg:%v\n", msg.MsgIDPop)

	}
	if ic := atomic.LoadInt32(&c.isClosed); ic > 0 {
		return nil
	}
	c.sendMsgChan <- info
	return nil
}

func (c *Clint) sendChangeRankMsg() error {
	info, ok := msg.MakeChangeRankMsg(c.UserId, c.prioritySecs)
	if !ok {
		return fmt.Errorf("sendChangeRankMsg.MakeChangeRankMsg error, msg:%v\n", msg.MsgIDPop)
	}
	if ic := atomic.LoadInt32(&c.isClosed); ic > 0 {
		return nil
	}
	c.sendMsgChan <- info
	return nil
}

func handleRankAckMsg(body []byte, c *Clint) {
	data := msg.MsgRankAck{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		c.Close()
		log.Printf("handleRankAckMsg err:%v\n", err)
		return
	}
	log.Printf("排队名次信息, 名次:%v, 用户ID:%v\n", data.Rank, c.UserId)
}

func handleOutMsg(body []byte, c *Clint) {
	c.Close()
	log.Printf("玩家排队结束,退出排队服务...\n")
}

func handleFullMsg(body []byte, c *Clint) {
	c.Close()
	log.Printf("队列已满, 请稍后重试...\n")
}

func handlePopAckMsg(body []byte, c *Clint) {
	data := msg.MsgQueuePopAck{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		c.Close()
		log.Printf("handlePopAckMsg err:%v\n", err)
		return
	}
	log.Printf("消费队列成功, 获得用户个数:%v...\n", data.Size)
}

func handleChangRankAckMsg(body []byte, c *Clint) {
	data := msg.MsgChangeRankAck{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		c.Close()
		log.Printf("handleChangRankAckMsg err:%v\n", err)
		return
	}
	log.Printf("调整玩家排队位置, 用户ID:%v, 调整前:%v, 调整后:%v...\n", data.UserId, data.OldRank, data.NewRank)
}
