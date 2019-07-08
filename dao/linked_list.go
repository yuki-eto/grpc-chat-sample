package dao

type LinkedList struct {
	head *LinkedListValue
	tail *LinkedListValue
	len  int
}

type LinkedListValue struct {
	Value interface{}
	next  *LinkedListValue
}

func NewLinkedList() *LinkedList {
	return new(LinkedList)
}

func (l *LinkedList) Add(value interface{}) {
	v := &LinkedListValue{
		Value: value,
	}
	if l.head == nil {
		l.head = v
		l.tail = v
	} else {
		l.tail.next = v
		l.tail = v
	}
	l.len++
}

func (l *LinkedList) AddHead(value interface{}) {
	v := &LinkedListValue{
		Value: value,
	}
	if l.head == nil {
		l.Add(value)
	} else {
		v.next = l.head
		l.head = v
	}
	l.len++
}

func (l *LinkedList) PopValue() *LinkedListValue {
	pop := l.head
	if pop == nil {
		return nil
	}

	l.head = l.head.next
	l.len--
	return pop
}

func (l *LinkedList) Each(f func(val *LinkedListValue)) {
	curr := l.head
	for {
		if curr == nil {
			return
		}
		f(curr)
		curr = curr.next
	}
}

func (l *LinkedList) Len() int {
	return l.len
}
