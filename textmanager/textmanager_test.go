package textmanager

import "testing"

func TestText(t *testing.T) {
    cursor := NewCursor()
    text := &Text{
        cursors: []*Cursor{cursor},
    }

    hello := "Hello, Wolrd!"

    for _, c := range hello {
        err := text.InsertCharBefore(0, c)
        if err != nil {
            t.Error(err)
        }
    }

    ptr := text.data.GetHead()
    println("fff")
    for ptr != nil {
        ptr1 := ptr.GetValue().GetHead()
        for ptr1 != nil {
            println(ptr1.GetValue())
            ptr1 = ptr1.GetNext()
        }

        ptr = ptr.GetNext()
    }
}
