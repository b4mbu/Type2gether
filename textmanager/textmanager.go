package textmanager

import (
    "Type2gether/list"
    "errors"
)

const (
    endl = '\n'
)

// type Line struct {
//     list list.List[rune]
// }
// Cursor указывает на текущий элемент, а не на злемент за ним
// Line теперь представляем не как наследуемую структуру от list.List[rune], а как list.List[rune] сам по себе. 
// чтобы создать новую строку, надо использовать line := new(list.List[rune])
// все вызовы остаются, как прежде


type Cursor struct {
    LineIter *list.Node[*list.List[rune]]
    CharIter *list.Node[rune]
    Id       int64
    Row      int32
    Col      int32
    Color    uint32
}

type Text struct {
    list.List[*list.List[rune]]
    // скорее всего не нужен тк в List есть аттрибут lenght /  size      int64
    Cursors   []*Cursor
}

func NewText() *Text{
    t := new(Text)
    t.PushBack(new(list.List[rune]))
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
       /* line := new(list.List[rune])
        err := line.PushBack(value)

        if err != nil {
            return err
        }

        err = t.PushBack(line)

        if err != nil {
            return err
        }

        cur.LineIter = t.GetHead()
        cur.CharIter = line.GetHead()
        cur.Col = 0
        cur.Row = 0

        return nil
        */
    }

    if cur.CharIter == nil {
        err := cur.LineIter.GetValue().PushFront(value)

        if err != nil {
            println(err)
            return err
        }

        cur.CharIter = cur.LineIter.GetValue().GetHead()
        cur.Col = 0
        return nil
    }

    err := cur.LineIter.GetValue().InsertAfter(value, cur.CharIter)

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

    if cur.CharIter == nil {
        line := new(list.List[rune])
        if cur.LineIter.GetValue().GetTail() != nil {
            // line isn't empty
            err := t.InsertBefore(line, cur.LineIter)

            if err != nil {
                return err
            }
            cur.Row++
            cur.Col = -1
            return nil
        }
        // line is empty
        err := t.InsertAfter(line, cur.LineIter)

        if err != nil {
            return err
        }

        cur.LineIter = cur.LineIter.GetNext()
        cur.CharIter = cur.LineIter.GetValue().GetHead()
        cur.Row++
        cur.Col = -1
        return nil
    }
    oldLineIter := cur.LineIter
    lst, err := oldLineIter.GetValue().Split(cur.CharIter.GetNext())

    if err != nil {
        return err
    }

    node := &list.Node[*list.List[rune]]{}
    node.SetValue(lst)

    node.SetPrev(oldLineIter)
    node.SetNext(oldLineIter.GetNext())
    if oldLineIter.GetNext() != nil {
        oldLineIter.GetNext().SetPrev(node)
    } else {
        t.SetTail(node)
    }
    oldLineIter.SetNext(node)
    println("Next:", cur.LineIter.GetNext()," = NODE:", node)
    cur.LineIter = node
    println("Head: ", cur.LineIter.GetValue().GetHead())
    cur.CharIter = nil

    cur.Row++
    cur.Col = -1
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
        oldTail := prev.GetValue().GetTail()
        oldLen := prev.GetValue().Length()
        err := t.MergeLines(cursorId)
        if err != nil {
            return err
        }
        cur.LineIter = prev
        cur.CharIter = oldTail
        cur.Row--
        cur.Col = oldLen - 1

        return nil
    }

    newCharIter := cur.CharIter.GetPrev()
    err := cur.LineIter.GetValue().Remove(cur.CharIter)

    if err != nil {
        return err
    }

    cur.CharIter = newCharIter
    cur.Col--


    return nil
}

// курсор не двигается!!
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
    //println("Original str: ", str)
    return str
}

func (cur *Cursor) MoveLeft() {
    if cur.CharIter == nil {
        if cur.LineIter.GetPrev() == nil {
            return
        }

        cur.LineIter = cur.LineIter.GetPrev()
        cur.CharIter = cur.LineIter.GetValue().GetTail()
        cur.Row--
        cur.Col = cur.LineIter.GetValue().Length() - 1
        return
    }

    cur.CharIter = cur.CharIter.GetPrev()
    cur.Col--
}

func (cur *Cursor) MoveRight() {
    if cur.CharIter == nil {
        if cur.LineIter.GetValue().GetHead() == nil {
            if cur.LineIter.GetNext() == nil {
                return
            }
            cur.LineIter = cur.LineIter.GetNext()
            cur.CharIter = nil
            cur.Row++
            cur.Col = -1
            return
        }

        cur.CharIter = cur.LineIter.GetValue().GetHead()
        cur.Col = 0
        return
    }

    if cur.CharIter == cur.LineIter.GetValue().GetTail() {
        if cur.LineIter.GetNext() == nil {
            return
        }

        cur.LineIter = cur.LineIter.GetNext()
        cur.CharIter = nil
        cur.Row++
        cur.Col = -1
        return
    }

    cur.CharIter = cur.CharIter.GetNext()
    cur.Col++
}

func (cur *Cursor) MoveUp () {
    if cur.LineIter.GetPrev() == nil {
        return
    }

    index := cur.LineIter.GetValue().Index(cur.CharIter)
    cur.CharIter = cur.LineIter.GetPrev().GetValue().GetNodeByIndex(index)
    cur.LineIter = cur.LineIter.GetPrev()
    cur.Row--
    cur.Col = cur.LineIter.GetValue().Index(cur.CharIter)
}

func (cur *Cursor) MoveDown() {
    if cur.LineIter.GetNext() == nil {
        return
    }

    index := cur.LineIter.GetValue().Index(cur.CharIter)
    cur.CharIter = cur.LineIter.GetNext().GetValue().GetNodeByIndex(index)
    cur.LineIter = cur.LineIter.GetNext()
    cur.Row++
    cur.Col = cur.LineIter.GetValue().Index(cur.CharIter)
}

/*
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
