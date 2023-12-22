package collection

import (
	"math/rand"
)

const (
	MaxLevel = 32
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

type SortedSet[K comparable, S Number, T any] struct {
	header   *skipListNode[K, S, T]
	tail     *skipListNode[K, S, T]
	maxLevel int
	length   int
}

type skipListLevel[K comparable, S Number, T any] struct {
	forward *skipListNode[K, S, T]
	span    int
}

type skipListNode[K comparable, S Number, T any] struct {
	backward *skipListNode[K, S, T]
	levels   []*skipListLevel[K, S, T]
	key      K
	score    S
	val      T
}

func NewSortedSet[K comparable, S Number, T any]() *SortedSet[K, S, T] {
	header := &skipListNode[K, S, T]{
		levels: make([]*skipListLevel[K, S, T], MaxLevel),
	}
	for l := 0; l < MaxLevel; l++ {
		header.levels[l] = &skipListLevel[K, S, T]{}
	}
	return &SortedSet[K, S, T]{
		header: header,
		tail:   header,
	}
}

func (set *SortedSet[K, S, T]) Min() T {
	if set.length > 0 {
		return set.header.levels[0].forward.val
	}
	return *new(T)
}

func (set *SortedSet[K, S, T]) Max() T {
	if set.length > 0 {
		return set.tail.val
	}
	return *new(T)
}

// TODO: test case
func (set *SortedSet[K, S, T]) PopMax() T {
	if set.length > 0 {
		node := set.header
		tail := set.tail
		for i := len(set.tail.levels) - 1; i >= 0; i-- {
			for ; node.levels[i].forward != tail; node = node.levels[i].forward {
			}
			node.levels[i].forward = tail
			node.levels[i].span = 0

			if set.header.levels[i].forward == nil {
				set.maxLevel = i - 1
			}
		}
		set.tail = tail.backward
		set.length--
		return tail.val
	}
	return *new(T)
}

func (set *SortedSet[K, S, T]) TopMax(m, n int) []T {
	items := make([]T, 0, n-m)
	node := set.header
	span := 0
	for i := set.maxLevel - 1; i >= 0; i-- {
		for ; node.levels[i].forward != nil; node = node.levels[i].forward {
			if span+node.levels[i].span >= m {
				break
			}
			span += node.levels[i].span
		}
		if span+node.levels[i].span == m {
			break
		}
	}

	for i := m; i < n; i++ {
		if node.levels[0].forward != nil {
			items = append(items, node.levels[0].forward.val)
			node = node.levels[0].forward
		} else {
			break
		}
	}
	return items
}

func (set *SortedSet[K, S, T]) Insert(key K, score S, t T) {
	totalSpan := 0
	node := set.header

	nlevel := randomLevel()
	if nlevel > set.maxLevel {
		set.maxLevel = nlevel
	}

	nodes := make([]*skipListNode[K, S, T], set.maxLevel)
	spans := make([]int, set.maxLevel)

	newnode := &skipListNode[K, S, T]{
		key:    key,
		score:  score,
		val:    t,
		levels: make([]*skipListLevel[K, S, T], nlevel),
	}
	for l := 0; l < nlevel; l++ {
		newnode.levels[l] = &skipListLevel[K, S, T]{}
	}

	for l := set.maxLevel - 1; l >= 0; l-- {
		for ; node.levels[l].forward != nil; node = node.levels[l].forward {
			if score < node.levels[l].forward.score {
				break
			}
			totalSpan += node.levels[l].span
		}
		nodes[l], spans[l] = node, totalSpan
	}

	forward := nodes[0].levels[0].forward
	if forward != nil {
		forward.backward = newnode
	} else {
		set.tail = newnode
	}
	newnode.backward = nodes[0]

	for i := 0; i < set.maxLevel; i++ {
		original := nodes[i].levels[i].forward
		newnode.levels[i].forward = original
		nodes[i].levels[i].forward = newnode

		if i < nlevel {
			originalSpan := nodes[i].levels[i].span
			nodes[i].levels[i].span = totalSpan - spans[i] + 1
			newnode.levels[i].span = originalSpan + 1 - nodes[i].levels[i].span
		} else {
			nodes[i].levels[i].span++
		}
	}

	set.length++
}

func (set *SortedSet[K, S, T]) Remove(key K, score S) {
	nodes := make([]*skipListNode[K, S, T], set.maxLevel)
	exist := false
	node := set.header
	for i := set.maxLevel - 1; i >= 0; i-- {
		for ; node.levels[i].forward != nil; node = node.levels[i].forward {
			if node.levels[i].forward.key == key {
				exist = true
				break
			}
			if score < node.levels[i].forward.score {
				break
			}
		}
		nodes[i] = node
	}

	if !exist {
		return
	}

	for i := set.maxLevel - 1; i >= 0; i-- {
		if nodes[i].levels[i].forward != nil {
			span := nodes[i].levels[i].forward.levels[i].span
			nodes[i].levels[i].forward = nodes[i].levels[i].forward.levels[i].forward
			nodes[i].levels[i].span += span - 1
		} else {
			nodes[i].levels[i].span--
		}

		if set.header.levels[i].forward == nil {
			set.maxLevel = i - 1
		}
	}

	if nodes[0].levels[0].forward != nil {
		nodes[0].levels[0].forward.backward = nodes[0]
	} else {
		set.tail = nodes[0]
	}

	set.length--
}

func (set *SortedSet[K, S, T]) Get(key K, score S) T {
	nodes := make([]*skipListNode[K, S, T], set.maxLevel)
	exist := false
	node := set.header
	for i := set.maxLevel - 1; i >= 0; i-- {
		for ; node.levels[i].forward != nil; node = node.levels[i].forward {
			if node.levels[i].forward.key == key {
				exist = true
				break
			}
			if score < node.levels[i].forward.score {
				break
			}
		}
		nodes[i] = node
	}

	if !exist {
		return nodes[0].levels[0].forward.val
	} else {
		return *new(T)
	}
}

func (set *SortedSet[K, S, T]) Get1(key K) T {
	node := set.header
	for ; node.levels[0].forward != nil; node = node.levels[0].forward {
		if node.levels[0].forward.key == key {
			return node.levels[0].forward.val
		}
	}
	return *new(T)
}

func (set *SortedSet[K, S, T]) Len() int {
	return set.length
}

func randomLevel() int {
	threshold := 2147483647 / 4
	level := 1
	for rand.Int() < threshold {
		level += 1
	}
	if level < MaxLevel {
		return level
	}
	return MaxLevel
}
