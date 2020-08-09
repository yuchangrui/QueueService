package network

import (
	"QueueService/pkg/sortedset"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
	"time"
)

func (s *Session) addQueque() bool {
	if queue.Len() >= int64(capSize) {
		return false
	}
	time := time.Now().UnixNano()
	queue.Add(s.UserId, float64(time))

	UserMap.Store(s.UserId, s)
	return true
}

func (s *Session) quitQueque() {
	if ok := atomic.CompareAndSwapInt32(&s.isQuitQueue, 0, 1); ok {
		UserMap.Delete(s.UserId)
		queue.Remove(s.UserId)
	}
}

func (s *Session) queuePop(size int64) (int64, bool) {
	if size >= queue.Len() {
		size = queue.Len()
	}
	if size == 0 {
		return 0, true
	}
	items := queue.Range(0, size, false)
	queue.RemoveByRank(0, size)

	for _, tmp := range items {
		uinf, ok := UserMap.Load(tmp)
		if !ok {
			// log.Printf("user not find, uid:%v", tmp)
			continue
		}
		sPop := uinf.(*Session)
		UserMap.Delete(s.UserId)
		sPop.sendOutMsg()
		sPop.Stop()
	}
	return int64(len(items)), true
}

func (s *Session) getRank() (int64, bool) {
	return queue.GetRank(s.UserId, false)
}

func (s *Session) changeRank(userId string, secs int64) (int64, int64, bool) {
	var new, old int64
	var data *sortedset.Element   // 用户排名分数信息
	var oldScore, newSocre string // for log

	_, ok := UserMap.Load(userId)
	if !ok {
		// 用户未在队列
		fmt.Printf("Session.changeRank not find, userId:%v\n", userId)
		goto onErr
	}

	old, ok = queue.GetRank(userId, false)
	if !ok {
		fmt.Printf("Session.queue.GetRank error, userId:%v\n", userId)
		goto onErr
	}
	data, ok = queue.Get(userId)
	if !ok {
		fmt.Printf("Session.queue.Get error, userId:%v\n", userId)
		goto onErr
	}

	ok = queue.Add(userId, data.Score-float64(secs*1000000000))
	if !ok {
		fmt.Printf("Session.queue.Add error, userId:%v\n", userId)
		goto onErr
	}
	new, ok = queue.GetRank(userId, false)
	if !ok {
		fmt.Printf("Session.queue.GetRank2 error, userId:%v\n", userId)
		goto onErr
	}

	oldScore = strconv.FormatFloat(data.Score, 'f', -1, 64)
	newSocre = strconv.FormatFloat(data.Score-float64(secs*1000000000), 'f', -1, 64)
	log.Printf("Session.changeRank, userId:%v, oldScore:%v, newSocre:%v", s.UserId, oldScore, newSocre)
	return old, new, true

onErr:
	return -1, -1, false
}
