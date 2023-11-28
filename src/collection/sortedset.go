package collection

import "log"

type SortedSet[T comparable] struct {
	items []T
}

func NewSortedSet[T comparable]() *SortedSet[T] {
	return &SortedSet[T]{items: make([]T, 0)}
}

func (set *SortedSet[T]) First() T {
	log.Fatal("not implemented")
	return *new(T)
}

func (set *SortedSet[T]) Top(n int) []T {
	log.Fatal("not implemented")
	return nil
}

func (set *SortedSet[T]) Insert(t T) {
	log.Fatal("not implemented")
}

func (set *SortedSet[T]) Remove(t T) {
	log.Fatal("not implemented")
}
