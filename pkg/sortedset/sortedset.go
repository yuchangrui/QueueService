package sortedset

import (
	"strconv"
	"sync"
)

type SortedSet struct {
	mu       sync.RWMutex
	dict     map[string]*Element
	skiplist *skiplist
}

func Make() *SortedSet {
	return &SortedSet{
		dict:     make(map[string]*Element),
		skiplist: makeSkiplist(),
	}
}

/*
 * return: has inserted new node
 */
func (s *SortedSet) Add(member string, score float64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	element, ok := s.dict[member]
	s.dict[member] = &Element{
		Member: member,
		Score:  score,
	}
	if ok {
		if score != element.Score {
			s.skiplist.remove(member, score)
			s.skiplist.insert(member, score)
		}
	} else {
		s.skiplist.insert(member, score)
	}
	return true
}

func (s *SortedSet) Len() int64 {
	return int64(len(s.dict))
}

func (s *SortedSet) Get(member string) (element *Element, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	element, ok = s.dict[member]
	if !ok {
		return nil, false
	}
	return element, true
}

func (s *SortedSet) Remove(member string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.dict[member]
	if ok {
		s.skiplist.remove(member, v.Score)
		delete(s.dict, member)
		return true
	}
	return false
}

/**
 * get 0-based rank
 */
func (s *SortedSet) GetRank(member string, desc bool) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	element, ok := s.dict[member]
	if !ok {
		return -1, false
	}
	r := s.skiplist.getRank(member, element.Score)
	if desc {
		r = s.skiplist.length - r
	} else {
		r--
	}
	// log.Printf("SortedSet.GetRank, userId:%v, rank:%v", member, element.Score)
	return r + 1, true // 默认排序从0开始,需要+1
}

/**
 * traverse [start, stop), 0-based rank
 */
func (s *SortedSet) ForEach(start int64, stop int64, desc bool, consumer func(element *Element) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	size := int64(s.Len())
	if start < 0 || start >= size {
		panic("illegal start " + strconv.FormatInt(start, 10))
	}
	if stop < start || stop > size {
		panic("illegal end " + strconv.FormatInt(stop, 10))
	}

	// find start node
	var node *Node
	if desc {
		node = s.skiplist.tail
		if start > 0 {
			node = s.skiplist.getByRank(int64(size - start))
		}
	} else {
		node = s.skiplist.header.level[0].forward
		if start > 0 {
			node = s.skiplist.getByRank(int64(start + 1))
		}
	}

	sliceSize := int(stop - start)
	for i := 0; i < sliceSize; i++ {
		if !consumer(&node.Element) {
			break
		}
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
	}
}

/**
 * return [start, stop), 0-based rank
 * assert start in [0, size), stop in [start, size]
 */
func (s *SortedSet) Range(start int64, stop int64, desc bool) []*Element {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sliceSize := int(stop - start)
	slice := make([]*Element, sliceSize)
	i := 0
	s.ForEach(start, stop, desc, func(element *Element) bool {
		slice[i] = element
		i++
		return true
	})
	return slice
}

func (s *SortedSet) Count(min *ScoreBorder, max *ScoreBorder) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var i int64 = 0
	// ascending order
	s.ForEach(0, s.Len(), false, func(element *Element) bool {
		gtMin := min.less(element.Score) // greater than min
		if !gtMin {
			// has not into range, continue foreach
			return true
		}
		ltMax := max.greater(element.Score) // less than max
		if !ltMax {
			// break through score border, break foreach
			return false
		}
		// gtMin && ltMax
		i++
		return true
	})
	return i
}

func (s *SortedSet) ForEachByScore(min *ScoreBorder, max *ScoreBorder, offset int64, limit int64, desc bool, consumer func(element *Element) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// find start node
	var node *Node
	if desc {
		node = s.skiplist.getLastInScoreRange(min, max)
	} else {
		node = s.skiplist.getFirstInScoreRange(min, max)
	}

	for node != nil && offset > 0 {
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
		offset--
	}

	// A negative limit returns all elements from the offset
	for i := 0; (i < int(limit) || limit < 0) && node != nil; i++ {
		if !consumer(&node.Element) {
			break
		}
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
		gtMin := min.less(node.Element.Score) // greater than min
		ltMax := max.greater(node.Element.Score)
		if !gtMin || !ltMax {
			break // break through score border
		}
	}
}

/*
 * param limit: <0 means no limit
 */
func (s *SortedSet) RangeByScore(min *ScoreBorder, max *ScoreBorder, offset int64, limit int64, desc bool) []*Element {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit == 0 || offset < 0 {
		return make([]*Element, 0)
	}
	slice := make([]*Element, 0)
	s.ForEachByScore(min, max, offset, limit, desc, func(element *Element) bool {
		slice = append(slice, element)
		return true
	})
	return slice
}

func (s *SortedSet) RemoveByScore(min *ScoreBorder, max *ScoreBorder) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := s.skiplist.RemoveRangeByScore(min, max)
	for _, element := range removed {
		delete(s.dict, element.Member)
	}
	return int64(len(removed))
}

/*
 * 0-based rank, [start, stop)
 */
func (s *SortedSet) RemoveByRank(start int64, stop int64) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := s.skiplist.RemoveRangeByRank(start+1, stop+1)
	for _, element := range removed {
		delete(s.dict, element.Member)
	}
	return int64(len(removed))
}
