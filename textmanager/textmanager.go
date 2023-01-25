package textmanager

import (
    "Type2gether/list"
)

const (
    endl = '\n'
)

type Line struct {
    list.List[rune]
}

type Cursor struct {
    lineIter *list.Node[*Line]
    charIter *list.Node[rune]
    id      int64
    row     int32
    col     int32
}

type Text struct {
    list.List[*Line]
    size      uint32
    cursors   []*Cursor
}

func NewCursor() *Cursor {
    c := new(Cursor)
    return c
}

func (t *Text) InsertCharBefore(cursorId int64, value rune) error {
    cur := t.cursors[cursorId]

    if cur.lineIter == nil {
        t.InsertLineAfter(0)

        line := t.GetHead().GetValue()
        err := line.InsertBefore(value, nil)

        if err != nil {
            return err
        }

        return nil
    }

    if cur.charIter == nil {
        err := cur.lineIter.GetValue().InsertBefore(value, nil)

        if err != nil {
            return err
        }

        return nil
    }

    err := cur.lineIter.GetValue().InsertBefore(value, cur.charIter)

    if err != nil {
        return err
    }

    return nil
}

func (t *Text) InsertLineAfter(cursorId int64) error {
    cur := t.cursors[cursorId]
    line := &Line{}

    err := t.InsertAfter(line, cur.lineIter)

    if err != nil {
        return err
    }
    
    if cur.lineIter != nil {
        cur.lineIter = cur.lineIter.GetNext()
    } else {
        cur.lineIter = t.GetHead()
    }
    cur.charIter = nil

    return nil
}


/*
func (c *Cursor) SetLine (l *Line) {
    
    c.linePos = l
}

func (c *Cursor) InsertCharBefore (value rune) error {
    c.InsertCharAfter(value)
    return c.Right()
}

func (c *Cursor) InsertCharAfter(value rune) error {
    //fmt.Println("InsertCharAfter")
    char := CreateChar(value)
    if c.linePos == nil {
        c.InsertLineAfter()
        c.head = c.linePos
    }


    if c.charPos == nil {
        c.linePos.value = char
        c.charPos = char
        return nil
    }

    if c.linePos.size >= 120 {
        return errors.New("over bad( 2) Too long")
    }

    oldNextChar := c.charPos.next
    c.charPos.next = char
    char.prev = c.charPos
    char.next = oldNextChar

    if oldNextChar != nil {
        oldNextChar.prev = char
    }

    return nil
}

// TODO text size++ ?????????

func (c *Cursor) InsertLineAfter() error {
    //fmt.Println("InsertLineAfter")
    line := CreateLine()

    if c.linePos == nil {
        c.linePos = line
        return nil
    }

    oldNextLine := c.linePos.next
    c.linePos.next = line
    line.prev, line.next = c.linePos, oldNextLine
    if oldNextLine != nil {
        oldNextLine.prev = line
    }

    return nil
}

func (c *Cursor) Left() error {
    if c.linePos == nil || c.charPos == nil {
        return errors.New("bof(0)")
    }

    if c.charPos.prev == nil && c.linePos.prev == nil {
        return errors.New("bof(1)")
    }

    if c.charPos.prev == nil {
        c.Up()
        c.setPositionInCurrentLine(c.linePos.size - 1)
        return nil
    }

    c.charPos = c.charPos.prev
    return nil
}

func (c *Cursor) Up() error {
    if c.linePos == nil || c.charPos == nil {
        return errors.New("пупупупуп2")
    }

    if c.linePos.prev == nil {
        return errors.New("bof")
    }

    charInd, err := c.GetIndexInCurrentLine()

    if err != nil {
        return err
    }

    c.linePos = c.linePos.prev
    c.setPositionInCurrentLine(charInd)

    return nil
}

func (c *Cursor) Right() error {
    if c.linePos == nil || c.charPos == nil {
        return errors.New("пупупупуп")
    }

    if c.charPos.next == nil && c.linePos.next == nil {
        return errors.New("eof")
    }

    if c.charPos.next == nil {
        c.Down()
        c.setPositionInCurrentLine(0)
        return nil
    }

    c.charPos = c.charPos.next
    return nil
}

func (c *Cursor) Down() error {
    if c.linePos == nil || c.charPos == nil {
        return errors.New("пупупупуп2")
    }

    if c.linePos.next == nil {
        return errors.New("eof")
    }

    charInd, err := c.GetIndexInCurrentLine()

    if err != nil {
        return err
    }

    c.linePos = c.linePos.next
    c.setPositionInCurrentLine(charInd)

    return nil
}

func (c *Cursor) GetIndexInCurrentLine() (uint8, error) {
    //fmt.Println("GetIndexInCurrentLine")
    if c.linePos == nil || c.charPos == nil {
        return 0, errors.New("пупупупуп3")
    }

    var (
        ptr = c.linePos.value
        i uint8 = 0
    )

    for ;ptr != c.charPos; ptr = ptr.next {
        if ptr == nil {
            return 0, errors.New("invalid cursor state")
        }
        i++
    }

    return i, nil
}

func (c *Cursor) setPositionInCurrentLine(index uint8) error {
    //fmt.Println("setPositionInCurrentLine")
    if c.linePos == nil {
        c.charPos = nil
        return errors.New("пупупупуп4")
    }

    var (
        ptr = c.linePos.value
        i uint8 = 0
    )
    if (ptr == nil) {
        return errors.New("normaldы")
    }

    for ;ptr.next != nil && i != index; ptr = ptr.next {
        i++
    }

    c.charPos = ptr
    return nil
}

func CreateTextFromFile(reader *bufio.Reader) *Cursor{ // TODO: add error
    cursor := CreateCursor()
    //fmt.Println("cursor was created succ")
    for {
        textLine, err := reader.ReadString(endl)
        for _, c := range textLine {
            cursor.InsertCharBefore(c)
          // cursor.InsertCharAfter(c)
           //cursor.Right()
        }
        if err == io.EOF {
            return cursor
        }

        cursor.InsertLineAfter()
        cursor.Down()

    }
    return cursor
}

func (t *Text) GetTextAsLine() string {
    res := ""
    linePtr := t.head
    for linePtr != nil {
        charPtr := linePtr.value
        for charPtr != nil {
            res += string(charPtr.value)
            charPtr = charPtr.next
        }
        linePtr = linePtr.next
    }
    return res
}
*/
