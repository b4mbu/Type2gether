package textmanager

import (
    "testing"
    "fmt"
)

func TestText(t *testing.T) {
    cursor := NewCursor()
    text := &Text{
        cursors: []*Cursor{cursor},
    }

    hello := "Hello, Wolrd!~Privet, Mir~~Cu, Mir"
    // hello := "Hello, World!"

    for _, c := range hello {
        if c == '~' {
            println("fff")
            err := text.InsertLineAfter(0)
            
            if err != nil {
                t.Error(err)
            }
        } else {
            err := text.InsertCharBefore(0, c)
            if err != nil {
                t.Error(err)
            }
        }
    }

    ptr := text.GetHead()
    for ptr != nil {
        ptr1 := ptr.GetValue().GetHead()
        for ptr1 != nil {
            fmt.Print(string(ptr1.GetValue()))
            ptr1 = ptr1.GetNext()
        }
        fmt.Print("±")
        ptr = ptr.GetNext()
    }
}
