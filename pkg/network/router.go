package network

import (
	"QueueService/pkg/msg"
	"encoding/json"
	"log"
)

// 协议处理器
var mapHandleFunc = make(map[uint32]HandleFunc)

// 处理函数
type HandleFunc func(body []byte, s *Session)

// 注册处理函数
func regHandleFunc(msgId uint32, f HandleFunc) {
	mapHandleFunc[msgId] = f
}

func init() {
	regHandleFunc(msg.MsgIDJoin, HandleJoinMsg)
	regHandleFunc(msg.MsgIDRank, HandleRankMsg)
	regHandleFunc(msg.MsgIDPop, HandlePopMsg)
	regHandleFunc(msg.MsgIDChangeRank, HandleChangRankMsg)
	regHandleFunc(msg.MsgIDHeartBeat, HandleHeartBeatMsg)
}

func HandleJoinMsg(body []byte, s *Session) {
	data := msg.MsgJoin{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("HandleJoinMsg err:%v\n", err)
		return
	}
	s.UserId = data.UserId
	if ok := s.addQueque(); !ok {
		s.sendFullResp()
		log.Printf("队列已满,稍后重试, userId:%v\n", data.UserId)
	}
}

func HandleRankMsg(body []byte, s *Session) {
	data := msg.MsgQueryRank{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("HandleRankMsg err:%v\n", err)
		return
	}
	s.sendRankResp()
}

func HandlePopMsg(body []byte, s *Session) {
	data := msg.MsgQueuePop{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("HandlePopMsg err:%v\n", err)
		return
	}
	size, ok := s.queuePop(data.Size)
	if !ok {
		log.Printf("MsgIDPop err:%v\n", err)
		return
	}
	s.sendMsgPopAckResp(size)
}

func HandleChangRankMsg(body []byte, s *Session) {
	data := msg.MsgChangeRank{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("HandleChangRankMsg err:%v\n", err)
		return
	}
	old, new, ok := s.changeRank(data.UserId, data.Secs)
	if !ok {
		log.Printf("changeRank err:%v\n", err)
		s.sendMsgChangRankAckResp(data.UserId, -1, -1)
		return
	}
	s.sendMsgChangRankAckResp(data.UserId, old, new)
}

func HandleHeartBeatMsg(body []byte, s *Session) {
	// log.Printf("HandleHeartBeatMsg...\n")
}
