package msg

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

// msgId
const (
	MsgIDJoin          = iota + uint32(1) // 加入队列
	MsgIDRank                             // 请求当前排队位置
	MsgIDRankAck                          // 排队位置应答
	MsgIDOut                              // 排队结束
	MsgIDFull                             // 队列已满,稍后重试
	MsgIDPop                              // 后端服务接受队列用户
	MsgIDPopAck                           // 后端服务接受队列用户, 应答
	MsgIDChangeRank                       // 调整排队位置
	MsgIDChangeRankAck                    // 调整排队位置应答
	MsgIDHeartBeat                        // 心跳消息
)

// ------------------------------
// | DataLen=4 | MsgId=4 | Data |
// ------------------------------

var (
	headerLen int
)

func init() {
	tmp := MsgHeader{}
	headerLen = int(unsafe.Sizeof(tmp.BodyLen) + unsafe.Sizeof(tmp.MsgId))
}

type MsgHeader struct {
	BodyLen uint32 // body长度
	MsgId   uint32 // msgID
}

type Msg struct {
	MsgHeader
	Body []byte //消息体
}

// 创建消息
func NewMsgPackage(msgId uint32, data []byte) *Msg {
	return &Msg{
		MsgHeader: MsgHeader{
			BodyLen: uint32(len(data)),
			MsgId:   msgId,
		},
		Body: data,
	}
}

// 获取msg的头消息
func getHeadLen() int {
	return headerLen
}

// PackMsg 封包
func PackMsg(msg *Msg) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	//写dataLen
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.BodyLen); err != nil {
		return nil, err
	}

	//写msgID
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.MsgId); err != nil {
		return nil, err
	}

	//写data数据
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.Body); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

// UnPackMsg 解包
func UnPackMsg(data []byte, length int) (*Msg, int, error) {
	msg := &Msg{}
	if length < getHeadLen() {
		// 粘包
		return nil, 0, nil
	}

	recBuff := bytes.NewBuffer(data[:length])
	// 读dataLen
	if err := binary.Read(recBuff, binary.LittleEndian, &msg.BodyLen); err != nil {
		return nil, 0, err
	}
	// 读msgID
	if err := binary.Read(recBuff, binary.LittleEndian, &msg.MsgId); err != nil {
		return nil, 0, err
	}

	end := getHeadLen() + int(msg.BodyLen)
	if length < end {
		// 粘包
		return nil, 0, nil
	}

	if msg.BodyLen > 0 {
		// 读data数据
		msg.Body = make([]byte, int(msg.BodyLen))
		copy(msg.Body, data[getHeadLen():end])
		// log.Printf("body:%v, [%v]", string(msg.Body), msg.Body)
	}
	return msg, end, nil
}
