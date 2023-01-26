package textmanager

import (
    "Type2gether/list"
    "errors"
)

const (
    endl = '\n'
)

type Line struct {
    list list.List[rune]
}
//Cursor указывает на текущий элемент, а не на злемент за ним
type Cursor struct {
    LineIter *list.Node[*Line]
    CharIter *list.Node[rune]
    Id      int64
    Row     int32
    Col     int32
}

type Text struct {
    list.List[*Line]
    // скорее всего не нужен тк в List есть аттрибут lenght /  size      int64
    Cursors   []*Cursor
}

func NewText() *Text{
    t := new(Text)
    t.PushBack(new(Line))
    return t
}

//TODO написать присвоение id для курсора
//TODO на данном этапе при создании NewCursor просто всегда будем 0 в него передавать, когда будем писать серверную часть нужно будет определиться с созданием Id
func NewCursor(Id int64) *Cursor {
    c := new(Cursor)
    c.Row = 0
    c.Col = -1
    c.Id = Id
    return c
}

//TODO сделать красиво
//сделать откат строки, и ячейки если у нас не вставилась буква
// строку откатываем, только если мы одновременно создаём строку и ячейку
// иначе откатываем только ячейку
/*func (t *Text) InsertCharBefore(cursorId int64, value rune) error {
    cur := t.Cursors[cursorId]

    if cur.LineIter == nil {
        println("1111111111111111111111111111")
        err := t.InsertLineAfter(cursorId)

        if err != nil {
            return err
        }
        
        line := t.GetHead().GetValue()
        err = line.InsertBefore(value, nil)
        if err != nil {
            return err
        }
        cur.Col++
        return nil
    }

    if cur.CharIter == nil {
        println("22222222222222222222222222222222222222")
        err := cur.LineIter.GetValue().InsertBefore(value, nil)

        if err != nil {
            return err
        }
        cur.Col++
        return nil
    }
    println("33333333333333333333333333333333333333")

    err := cur.LineIter.GetValue().InsertBefore(value, cur.CharIter)

    if err != nil {
        return err
    }
    cur.Col++
    return nil
}*/

func (t *Text) InsertCharAfter(cursorId int64, value rune) error {
    cur := t.Cursors[cursorId]
    // эта штука не должна вызываться никогда
    if cur.LineIter == nil {
        println("Мужики, работяги, всё плохо. Юра, мы всё прое****, это условие не должно выполняться")
        line := new(Line)
        err := line.list.PushBack(value)

        if err != nil {
            return err
        }

        err = t.PushBack(line)

        if err != nil {
            return err
        }

        cur.LineIter = t.GetHead()
        cur.CharIter = line.list.GetHead()
        cur.Col = 0 
        cur.Row = 0

        return nil
    }

    if cur.CharIter == nil {
        err := cur.LineIter.GetValue().list.PushBack(value)

        if err != nil {
            return err
        }

        cur.CharIter = cur.LineIter.GetValue().list.GetHead()
        cur.Col = 0
        return nil
    }
    
    err := cur.LineIter.GetValue().list.InsertAfter(value, cur.CharIter)

    if err != nil {
        return err 
    }

    cur.CharIter = cur.CharIter.GetNext()
    cur.Col++
    return nil
}

func (t *Text) InsertLineAfter(cursorId int64) error {
    cur := t.Cursors[cursorId]

    if cur.LineIter == nil {
        println("Мы не должны заходить в этот if ")
        return errors.New("Мы не должны заходить в этот if")
    }
    
    line := new(Line)

    if cur.CharIter == nil {
        err := t.InsertAfter(line, cur.LineIter)

        if err != nil {
            return err    
        }

        cur.LineIter = cur.LineIter.GetNext()
        cur.CharIter = cur.LineIter.GetValue().list.GetHead()
        cur.Row++
        cur.Col = -1
        return nil
    }
    
    lst, err := cur.LineIter.GetValue().list.Split(cur.CharIter)
    
    if err != nil {
        return err
    }

    node := &list.Node[*list.List[rune]]{}
    node.SetValue(lst)
    
    node.SetPrev(cur.LineIter.list)
    node.SetNext(cur.LineIter.GetNext())
    if cur.LineIter.GetNext() == nil {
        cur.LineIter.GetNext().SetPrev(node)
    }
    cur.LineIter.SetNext(node)
    cur.LineIter = node
    //cur.CharIter = node.GetValue().GetHead() если это не работает, то Егор не прав но, только в этом моменте
    cur.Row++
    cur.Col = 0
    return nil
}

//TODO cur.CharIter == nil не значит, что строка пустая || Нужно проверить везде в коде это 
func (t *Text) RemoveCharBefore(cursorId int64) error {
    cur := t.Cursors[cursorId]
    
    if cur.LineIter == nil {
        println("Press 'F', пацан к успеху шёл, Какой успех, раздался смех")
        return nil
    }
    
    if cur.CharIter == nil {
        if cur.LineIter.GetPrev() == nil {
            return nil
        }
        prev := cur.LineIter.GetPrev()
        err := t.MergeLines(cursorId)
        if err != nil {
            return err
        }
        cur.LineIter = prev
        cur.CharIter = cur.LineIter.GetValue().GetTail()
        return nil
    }
    
    return nil
}

func (t *Text) MergeLines(cursorId int64) error {
    cur := t.Cursors[cursorId]
    if cur.LineIter.GetPrev() == nil {
        return errors.New("Нет предыдущей строки")
    }
    err := cur.LineIter.GetPrev().GetValue().Merge(cur.LineIter.GetValue())
    if err != nil {
        return err
    }
    return t.Remove(cur.LineIter)
}

func (t *Text) GetString() string {
    str := ""
    iter := t.GetHead()
    for iter != nil {
        char_iter := iter.GetValue().GetHead()
        for char_iter != nil {
            str += string(char_iter.GetValue())
            char_iter = char_iter.GetNext()
        }
        str += "\n"
        iter = iter.GetNext()
    }
    return str
}
/*
func (c *Cursor) SetLine (l *Line) {
    
    c.linePos = l
}
f cur.LineIter != nil {
 92         cur.LineIter 
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
