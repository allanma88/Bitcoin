package collection

//TODO: test cases

type ListMap[K bool | int | string, T any] struct {
	head  *listNode[K, T]
	tail  *listNode[K, T]
	nodes map[K]*listNode[K, T]
}

type listNode[K bool | int | string, T any] struct {
	Key  K
	Val  T
	Prev *listNode[K, T]
	Next *listNode[K, T]
}

func NewListMap[K bool | int | string, T any]() *ListMap[K, T] {
	list := &ListMap[K, T]{
		nodes: make(map[K]*listNode[K, T]),
	}
	return list
}

func (list *ListMap[K, T]) FirstKey() K {
	if list.head != nil {
		return list.head.Key
	}
	return *new(K)
}

func (list *ListMap[K, T]) NextKey(key K) K {
	if node, ok := list.nodes[key]; ok {
		if node.Next != nil {
			return node.Next.Key
		}
	}
	return *new(K)
}

func (list *ListMap[K, T]) Get(key K) T {
	if node, ok := list.nodes[key]; ok {
		return node.Val
	} else {
		return *new(T)
	}
}

func (list *ListMap[K, T]) Len() int {
	return len(list.nodes)
}

func (list *ListMap[K, T]) Keys() []K {
	keys := make([]K, 0, len(list.nodes))
	for k := range list.nodes {
		keys = append(keys, k)
	}
	return keys
}

func (list *ListMap[K, T]) Set(key K, val T) {
	//TODO: in progress
	if _, has := list.nodes[key]; !has {
		node := &listNode[K, T]{Key: key, Val: val}

		if list.head == nil && list.tail == nil {
			list.head = node
			list.tail = node
		} else {
			node.Prev = list.tail
			node.Next = list.head

			list.tail.Next = node
			list.head.Prev = node

			list.tail = node
		}
		list.nodes[key] = node
	} else {
		list.nodes[key].Val = val
	}
}

func (list *ListMap[K, T]) Remove(key K) {
	//TODO: in progress
	node := list.nodes[key]

	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev

	if node == list.tail {
		list.tail = node.Prev
	} else if node == list.head {
		list.head = node.Next
	}

	node.Prev = nil
	node.Next = nil

	delete(list.nodes, key)
}
