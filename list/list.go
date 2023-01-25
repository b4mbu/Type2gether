// TODO TEST EVERYTHING!!!
package list

import (
    "fmt"
    "errors"
)

type List[T any] struct {
    head   *Node[T]
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

func (l *List[T]) PushBack(value T) {
    if l.head == nil {
        l.head = NewNode(value)
        l.length = 1
        return
    }
    ptr := l.head
    for ptr.next != nil {
        ptr = ptr.next
    }
    node := NewNode(value)
    ptr.next = node
    node.prev = ptr
    l.length++
}

func (l *List[T]) InsertBefore(value T, node *Node[T]) error {
    if l.head == nil && node == nil {
        l.head = NewNode(value)
        l.length = 1
        return nil
    }

    if l.head == nil {
        return errors.New("head is nil but node is not nil")
    }

    fmt.Println(node)
    fmt.Println(l.head)
    if node == nil {
        l.PushBack(value)
        return nil
    }

    if node == l.head {
        newNode := NewNode(value)
        l.head.prev = newNode
        newNode.next = l.head
        l.head = newNode
        l.length++
        return nil
    }

    ptr := l.head
    for ptr != nil && ptr.next != node {
        ptr = ptr.next
    }

    if ptr == nil {
        l.PushBack(value)
        return nil
    }

    newNode := NewNode(value)
    ptr.next = newNode
    newNode.prev = ptr
    newNode.next = node
    node.prev = newNode
    l.length++
    return nil
}

func (l *List[T]) InsertAfter(value T, node *Node[T]) error {
    if l.head == nil && node == nil {
        l.head = NewNode(value)
        l.length = 1
        return nil
    }

    if l.head == nil {
        return errors.New("head is nil but node is not nil")
    }

    if node == nil {
        l.PushBack(value)
        return nil
    }

    ptr := l.head

    for ptr != nil && ptr != node {
        ptr = ptr.next
    }

    if ptr == nil {
        return errors.New("target node not in list")
    }

    var (
        newNode = NewNode(value)
        oldNext = ptr.next
    )

    ptr.next = newNode
    newNode.prev = ptr
    newNode.next = oldNext
    if oldNext != nil {
        oldNext.prev = newNode
    }
    l.length++
    return nil
}

//TODO: fix Remove with l.head == nil
func (l *List[T]) Remove(node *Node[T]) {
    if node == nil || l.head == nil {
        return
    }

    if node == l.head {
        l.length = 0
        l.head = l.head.next
        if l.head == nil {
            return
        }

        l.head.prev = nil
        return
    }

    ptr := l.head
    for ptr != nil && ptr != node {
        ptr = ptr.next
    }

    if ptr == nil {
        return
    }

    if ptr.next == nil {
        ptr.prev.next = nil
        l.length--
        return
    }

    ptr.prev.next = ptr.next
    ptr.next.prev = ptr.prev
    l.length--
}


func (l *List[T]) GetHead() *Node[T] {
    return l.head
}

func (l *List[T]) Length() int32 {
    return l.length
}

