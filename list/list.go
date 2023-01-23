// TODO TEST EVERYTHING!!!
package list

import (
    //"fmt"
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

func (l *List[T]) InsertBefore(value T, node *Node[T]) {
    if l.head == nil {
        return
    }

    if node == nil {
        l.PushBack(value)
        return
    }

    ptr := l.head
    for ptr.next != node {
        ptr = ptr.next
    }

    newNode := NewNode(value)
    ptr.next = newNode
    newNode.prev = ptr
    newNode.next = node
    node.prev = newNode
    l.length++
}

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
    for ptr != node {
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

func main() {

}
