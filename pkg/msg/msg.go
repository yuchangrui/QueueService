package msg

import (
	"encoding/json"
	"log"
)

// MsgJoin 加入队列
type MsgJoin struct {
	UserId string `json:"userId"`
}

// MsgQueryRank 请求排名
type MsgQueryRank struct {
	UserId string `json:"userId"`
}

// MsgIDRankAck 请求排名应答
type MsgRankAck struct {
	UserId string `json:"userId"`
	Rank   int64  `json:"rank"`
}

// MsgQueuePop 消费队列
type MsgQueuePop struct {
	Size int64 `json:"size"`
}

// MsgQueuePop 消费队列应答
type MsgQueuePopAck struct {
	Size int64 `json:"size"`
}

// MsgQueuePop 消费队列应答
type MsgChangeRank struct {
	UserId string `json:"userId"`
	Secs   int64  `json:"secs"`
}

// MsgQueuePop 消费队列应答
type MsgChangeRankAck struct {
	UserId  string `json:"userId"`
	OldRank int64  `json:"oldRank"`
	NewRank int64  `json:"newRank"`
}

func MakeJoinMsg(userId string) ([]byte, bool) {
	tmp := MsgJoin{UserId: userId}
	body, _ := json.Marshal(tmp)
	ret, err := PackMsg(NewMsgPackage(MsgIDJoin, body))
	if err != nil {
		log.Printf("msg.MakeJsonMsg PackMsg error, userId:%v, err:%v\n", userId, err)
		return ret, false
	}
	return ret, true
}

func MakeRankMsg(userId string) ([]byte, bool) {
	tmp := MsgQueryRank{UserId: userId}
	body, _ := json.Marshal(tmp)
	ret, err := PackMsg(NewMsgPackage(MsgIDRank, body))
	if err != nil {
		log.Printf("msg.MakeRankMsg PackMsg error, userId:%v, err:%v\n", userId, err)
		return ret, false
	}
	return ret, true
}

func MakeRankAckMsg(userId string, rank int64) ([]byte, bool) {
	tmp := MsgRankAck{UserId: userId, Rank: rank}
	body, _ := json.Marshal(tmp)
	ret, err := PackMsg(NewMsgPackage(MsgIDRankAck, body))
	if err != nil {
		log.Printf("msg.MakeRankAckMsg PackMsg error, userId:%v, err:%v\n", userId, err)
		return ret, false
	}
	return ret, true
}

func MakeOutMsg() ([]byte, bool) {
	ret, err := PackMsg(NewMsgPackage(MsgIDOut, []byte{}))
	if err != nil {
		log.Printf("msg.MakeOutMsg PackMsg error:%v\n", err)
		return ret, false
	}
	return ret, true
}

func MakePopMsg(size int64) ([]byte, bool) {
	tmp := MsgQueuePop{Size: size}
	body, _ := json.Marshal(tmp)
	ret, err := PackMsg(NewMsgPackage(MsgIDPop, body))
	if err != nil {
		log.Printf("msg.MakePopMsg PackMsg error, err:%v\n", err)
		return ret, false
	}
	return ret, true
}

func MakePopAckMsg(size int64) ([]byte, bool) {
	tmp := MsgQueuePopAck{Size: size}
	body, _ := json.Marshal(tmp)
	ret, err := PackMsg(NewMsgPackage(MsgIDPopAck, body))
	if err != nil {
		log.Printf("msg.MakePopAckMsg PackMsg error, err:%v\n", err)
		return ret, false
	}
	return ret, true
}

func MakeFullMsg() ([]byte, bool) {
	ret, err := PackMsg(NewMsgPackage(MsgIDFull, []byte{}))
	if err != nil {
		log.Printf("msg.MakeOutMsg PackMsg error:%v\n", err)
		return ret, false
	}
	return ret, true
}

func MakeChangeRankMsg(userId string, secs int64) ([]byte, bool) {
	tmp := MsgChangeRank{UserId: userId, Secs: secs}
	body, _ := json.Marshal(tmp)
	ret, err := PackMsg(NewMsgPackage(MsgIDChangeRank, body))
	if err != nil {
		log.Printf("msg.MakeChangeRankMsg PackMsg error, err:%v\n", err)
		return ret, false
	}
	return ret, true
}

func MakeChangeRankAckMsg(userId string, old int64, new int64) ([]byte, bool) {
	tmp := MsgChangeRankAck{UserId: userId, OldRank: old, NewRank: new}
	body, _ := json.Marshal(tmp)
	ret, err := PackMsg(NewMsgPackage(MsgIDChangeRankAck, body))
	if err != nil {
		log.Printf("msg.MakeChangeRankAckMsg PackMsg error, err:%v\n", err)
		return ret, false
	}
	return ret, true
}

func MakeHeartBeatMsg() ([]byte, bool) {
	ret, err := PackMsg(NewMsgPackage(MsgIDHeartBeat, []byte{}))
	if err != nil {
		log.Printf("msg.MakeHeartBeatMsg PackMsg error:%v\n", err)
		return ret, false
	}
	return ret, true
}
