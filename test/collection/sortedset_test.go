package collection

import (
	"Bitcoin/src/collection"
	"testing"
)

func Test_Insert_Max(t *testing.T) {
	sortedset := collection.NewSortedSet[*Entity]()
	n := 10
	for i := 0; i < n; i++ {
		sortedset.Insert(&Entity{id: i + 1, number: i + 1})

		max := sortedset.Max()
		if max == nil {
			t.Fatal("no max entity after insert")
		}
		if sortedset.Len() != i+1 {
			t.Fatalf("expect len: %d, actual: %d", i+1, sortedset.Len())
		}
		if max.id != i+1 || max.number != i+1 {
			t.Fatalf("expect max: %d-%d, actual: %d-%d", i+1, i+1, max.id, max.number)
		}
	}
}

func Test_Delete_Max(t *testing.T) {
	sortedset := collection.NewSortedSet[*Entity]()
	n := 10
	for i := 0; i < n; i++ {
		sortedset.Insert(&Entity{id: i + 1, number: i + 1})
	}

	for i := n - 1; i >= 0; i-- {
		max := sortedset.Max()
		if max == nil {
			t.Fatal("no max entity when remove")
		}
		if sortedset.Len() != i+1 {
			t.Fatalf("expect len: %d, actual: %d", i+1, sortedset.Len())
		}
		if max.id != i+1 || max.number != i+1 {
			t.Fatalf("expect max: %d-%d, actual: %d-%d", i+1, i+1, max.id, max.number)
		}
		sortedset.Remove(max)
	}

	if sortedset.Len() != 0 {
		t.Fatalf("expect len: %d, actual: %d", 0, sortedset.Len())
	}
}

func Test_Insert_Top_Max(t *testing.T) {
	sortedset := collection.NewSortedSet[*Entity]()
	n, batch := 100, 10
	for i := 0; i < n; i++ {
		sortedset.Insert(&Entity{id: i + 1, number: i + 1})
	}

	entites := make([]*Entity, 0, n)
	for s := 0; s < n; s += batch {
		entites1 := sortedset.TopMax(s, s+batch)
		if len(entites1) != batch {
			t.Fatalf("expect entities len: %d, actual: %d", batch, len(entites))
		}
		entites = append(entites, entites1...)
	}

	if len(entites) != n {
		t.Fatalf("expect all entities len: %d, actual: %d", n, len(entites))
	}

	for i := 1; i < n; i++ {
		entity1, entity2 := entites[i-1], entites[i]
		if entity1.number > entity2.number {
			t.Fatalf("The %d entity is smaller than the %d entity", i, i+1)
		}
	}
}

func Test_Empty_Max(t *testing.T) {
	sortedset := collection.NewSortedSet[*Entity]()
	max := sortedset.Max()
	if max != nil {
		t.Fatal("get max entity for empty set")
	}
	if sortedset.Len() != 0 {
		t.Fatalf("expect len: %d, actual: %d", 0, sortedset.Len())
	}
}

func Test_Empty_Remove(t *testing.T) {
	sortedset := collection.NewSortedSet[*Entity]()
	sortedset.Remove(&Entity{})
}

type Entity struct {
	id     int
	number int
}

func (block *Entity) Compare(other collection.Comparable) int {
	otherBlock := other.(*Entity)
	if block.number < otherBlock.number {
		return -1
	} else if block.number == otherBlock.number {
		return 0
	}
	return 1
}

func (block *Entity) Equal(other collection.Comparable) bool {
	otherBlock := other.(*Entity)
	return block.id == otherBlock.id
}
