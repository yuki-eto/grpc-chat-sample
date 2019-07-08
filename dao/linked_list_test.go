package dao

import (
	"testing"
)

func TestLinkedList_PopValue(t *testing.T) {
	linkedList := NewLinkedList()

	for i := 1; i <= 5; i++ {
		linkedList.Add(i)
	}
	linkedList.AddHead(100)
	for i := 6; i <= 10; i++ {
		linkedList.Add(i)
	}

	if linkedList.Len() != 11 {
		t.Fatalf("count: %d", linkedList.Len())
	}

	l := linkedList.len
	assertList := []int{100, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 0; i < l; i++ {
		val := linkedList.PopValue()
		if assertList[i] != val.Value.(int) {
			t.Fatalf("not match: %d <> %d", assertList[i], val.Value)
		}
	}
}
