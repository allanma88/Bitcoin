package collection

import (
	"math/rand"
)

const (
	MaxLevel = 32
)

type Comparable interface {
	Compare(another Comparable) int
	Equal(another Comparable) bool
}

type SortedSet[T Comparable] struct {
	header   *skipListNode[T]
	tail     *skipListNode[T]
	maxLevel int
	length   int
}

type skipListLevel[T Comparable] struct {
	forward *skipListNode[T]
	span    int
}

type skipListNode[T Comparable] struct {
	backward *skipListNode[T]
	levels   []*skipListLevel[T]
	val      T
}

func NewSortedSet[T Comparable]() *SortedSet[T] {
	header := &skipListNode[T]{
		levels: make([]*skipListLevel[T], MaxLevel),
	}
	for l := 0; l < MaxLevel; l++ {
		header.levels[l] = &skipListLevel[T]{}
	}
	return &SortedSet[T]{
		header: header,
		tail:   header,
	}
}

func (set *SortedSet[T]) Min() T {
	if set.length > 0 {
		return set.header.levels[0].forward.val
	}
	return *new(T)
}

func (set *SortedSet[T]) Max() T {
	if set.length > 0 {
		return set.tail.val
	}
	return *new(T)
}

func (set *SortedSet[T]) TopMax(m, n int) []T {
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

func (set *SortedSet[T]) Insert(t T) {
	totalSpan := 0
	node := set.header

	nlevel := randomLevel()
	if nlevel > set.maxLevel {
		set.maxLevel = nlevel
	}

	nodes := make([]*skipListNode[T], set.maxLevel)
	spans := make([]int, set.maxLevel)

	newnode := &skipListNode[T]{
		val:    t,
		levels: make([]*skipListLevel[T], nlevel),
	}
	for l := 0; l < nlevel; l++ {
		newnode.levels[l] = &skipListLevel[T]{}
	}

	for l := set.maxLevel - 1; l >= 0; l-- {
		for ; node.levels[l].forward != nil; node = node.levels[l].forward {
			if t.Compare(node.levels[l].forward.val) < 1 {
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

func (set *SortedSet[T]) Remove(t T) {
	nodes := make([]*skipListNode[T], set.maxLevel)
	exist := false
	node := set.header
	for i := set.maxLevel - 1; i >= 0; i-- {
		for ; node.levels[i].forward != nil; node = node.levels[i].forward {
			if node.levels[i].forward.val.Equal(t) {
				exist = true
				break
			}
			if t.Compare(node.levels[i].forward.val) < 1 {
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

func (set *SortedSet[T]) Len() int {
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
