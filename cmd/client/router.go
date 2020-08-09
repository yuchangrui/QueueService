package main

import (
	"QueueService/pkg/msg"
)

// 全局协议处理器
var mapHandleFunc = make(map[uint32]HandleFunc)

// 处理器格式
type HandleFunc func(body []byte, c *Clint)

// 注册处理器
func regHandleFunc(msgId uint32, f HandleFunc) {
	mapHandleFunc[msgId] = f
}

// 初始化
func init() {
	regHandleFunc(msg.MsgIDRankAck, handleRankAckMsg)
	regHandleFunc(msg.MsgIDOut, handleOutMsg)
	regHandleFunc(msg.MsgIDFull, handleFullMsg)
	regHandleFunc(msg.MsgIDPopAck, handlePopAckMsg)
	regHandleFunc(msg.MsgIDChangeRankAck, handleChangRankAckMsg)
}
