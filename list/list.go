// TODO TEST EVERYTHING!!!
package list

import (
	"errors"
)

type List[T any] struct {
	head   *Node[T]
	tail   *Node[T]
	length int32
}

type Node[T any] struct {
	value T
	next  *Node[T]
	prev  *Node[T]
}

func NewNode[T any](value T) *Node[T] {
	return &Node[T]{value: value}
}

func (n *Node[T]) GetNext() *Node[T] {
	return n.next
}

func (n *Node[T]) GetPrev() *Node[T] {
	return n.prev
}

func (n *Node[T]) GetValue() T {
	return n.value
}

func (n *Node[T]) SetValue(value T) {
	n.value = value
}

func (l *List[T]) SetHead(value *Node[T]) {
	l.head = value
}

func (l *List[T]) SetTail(value *Node[T]) {
	l.tail = value
}

func (n *Node[T]) SetNext(value *Node[T]) {
	n.next = value
}

func (n *Node[T]) SetPrev(value *Node[T]) {
	n.prev = value
}

func (l *List[T]) InsertBefore(value T, node *Node[T]) error {
	if l.head == nil && node == nil && l.tail == nil {
		l.head = NewNode(value)
		l.length = 1
		l.tail = l.head
		return nil
	}

	if l.head == nil {
		return errors.New("head is nil but node is not nil")
	}

	if l.tail == nil {
		return errors.New("tail is nil")
	}

	if node == nil {
		return errors.New("node is nil")
	}

	if node == l.head {
		return l.PushFront(value)
	}

	if !l.Consists(node) {
		return errors.New("node not in list")
	}

	var (
		newNode = NewNode(value)
		oldPrev = node.prev
	)

	node.prev = newNode
	newNode.next = node
	newNode.prev = oldPrev

	if oldPrev != nil {
		oldPrev.next = newNode
	}

	l.length++
	return nil
}

func (l *List[T]) PushFront(value T) error {
	if l.head == nil && l.tail == nil {
		l.head = NewNode(value)
		l.length = 1
		l.tail = l.head
		return nil
	}

	if l.tail == nil {
		return errors.New("tail is nil")
	}
	if l.head == nil {
		return errors.New("head is nil")
	}

	node := NewNode(value)
	l.head.prev = node
	node.next = l.head
	l.head = node
	l.length++
	return nil
}

func (l *List[T]) InsertAfter(value T, node *Node[T]) error {
	if l.head == nil && node == nil && l.tail == nil {
		l.head = NewNode(value)
		l.length = 1
		l.tail = l.head
		return nil
	}

	if l.head == nil {
		return errors.New("head is nil but node is not nil")
	}

	if l.tail == nil {
		return errors.New("tail is nil")
	}

	if node == nil {
		return errors.New("node is nil")
		//return l.PushBack(value)
	}

	if node == l.tail {
		return l.PushBack(value)
	}

	if !l.Consists(node) {
		return errors.New("node not in list")
	}

	var (
		newNode = NewNode(value)
		oldNext = node.next
	)

	node.next = newNode
	newNode.prev = node
	newNode.next = oldNext
	if oldNext != nil {
		oldNext.prev = newNode
	}
	l.length++
	return nil
}

func (l *List[T]) PushBack(value T) error {
	if l.head == nil && l.tail == nil {
		l.head = NewNode(value)
		l.length = 1
		l.tail = l.head
		return nil
	}

	if l.tail == nil {
		return errors.New("tail is nil")
	}
	if l.head == nil {
		return errors.New("head is nil")
	}

	node := NewNode(value)
	l.tail.next = node
	node.prev = l.tail
	l.tail = node
	l.length++
	return nil
}

//TODO: fix Remove with l.head == nil
func (l *List[T]) Remove(node *Node[T]) error {
	if node == nil || l.head == nil || l.tail == nil {
		return errors.New("node or head or tail are nil")
	}

	if node == l.head {
		l.length--
		l.head = l.head.next
		if l.head == nil {
			l.tail = nil
			return nil
		}

		l.head.prev = nil
		return nil
	}
	// Если prev tail == nil, то tail == head, а это мы проверили выше
	if node == l.tail {
		l.tail = l.tail.prev
		l.tail.next = nil
		l.length--
		return nil
	}

	if !l.Consists(node) {
		return errors.New("node not in list")
	}

	if node.next == nil || node.prev == nil {
		return errors.New("node dose not have next or tail")
	}

	node.prev.next = node.next
	node.next.prev = node.prev
	l.length--
	return nil
}

func (l *List[T]) GetHead() *Node[T] {
	return l.head
}

func (l *List[T]) GetTail() *Node[T] {
	return l.tail
}

func (l *List[T]) Length() int32 {
	return l.length
}

func (l *List[T]) Consists(node *Node[T]) bool {
	if node == nil {
		return false
	}

	ptr := l.head

	for ptr != nil && ptr != node {
		ptr = ptr.next
	}

	return ptr == node
}

func (l *List[T]) Index(node *Node[T]) int32 {
	if node == nil {
		return -1
	}

	ptr := l.head
	var ind int32

	for ptr != nil && ptr != node {
		ptr = ptr.next
		ind += 1
	}

	if ptr == node {
		return ind
	}
	return -1
}

func (l *List[T]) GetNodeByIndex(index int32) *Node[T] {
	if index == -1 {
		return nil
	}

	if index >= l.length {
		return l.tail
	}
	ptr := l.head
	for index != 0 {
		index--
		ptr = ptr.next
	}
	return ptr
}

// split принимает ту ноду, которая станет головой
func (l *List[T]) Split(node *Node[T]) (*List[T], error) {
	if node == nil {
		return &List[T]{}, nil
	}

	if !l.Consists(node) {
		return nil, errors.New("node not in list")
	}

	list := new(List[T])
	list.head = node
	list.tail = l.tail

	if node.prev == nil {
		l.head = nil
		l.tail = nil
		list.length = l.length
		l.length = 0
		return list, nil
	}

	list.length = l.length - l.Index(node)
	l.length -= list.length
	l.tail = node.prev
	l.tail.next = nil
	list.head.prev = nil
	return list, nil
}

//TODO доделать merge на уровне текста
func (l *List[T]) Merge(r *List[T]) error {
	if r == nil {
		return errors.New("list is nil")
	}

	if l.head == nil && l.tail == nil {
		l.head = r.head
		l.tail = r.tail
		l.length = r.length
		r = nil
		return nil
	}

	if l.head == nil {
		return errors.New("left list in merge: tail is not nil where head is nil")
	}

	if l.tail == nil {
		return errors.New("left list in merge: head is not nil where tail is nil")
	}

	if r.head == nil && r.tail == nil {
		return nil
	}

	if r.head == nil {
		return errors.New("right list in merge: tail is not nil where head is nil ")
	}

	if r.tail == nil {
		return errors.New("right list in merge: head is not nil where tail is nil")
	}
	l.tail.next = r.head
	r.head.prev = l.tail
	l.tail = r.tail
	l.length += r.length
	r = nil
	return nil
}

func (l *List[T]) CopyTo(other *List[T]) {
	if other == nil {
		return
	}

	other.head = l.head
	other.tail = l.tail
	other.length = l.length
}

func (l *List[T]) Clear() {
	l.head, l.tail, l.length = nil, nil, 0
}
