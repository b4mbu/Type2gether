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
	LineIter   *list.Node[*list.List[rune]]
	CharIter   *list.Node[rune]
	Id         int64
	Row        int32
	Col        int32
	Color      uint32
	ScreenHead *Border
	ScreenTail *Border
	ScreenRow  int32
}

//Эта структура нужна для отрисовки нумерации строк
type Border struct {
	LineIter  *list.Node[*list.List[rune]]
	RowNumber int32
}

type Text struct {
	list.List[*list.List[rune]]
	// скорее всего не нужен тк в List есть аттрибут lenght /  size      int64
	Cursors []*Cursor
}

func NewText() *Text {
	t := new(Text)
	t.PushBack(new(list.List[rune]))
	// for _, cur := range t.Cursors {
	// 	cur.ScreenHead.LineIter = t.GetHead()
	// 	cur.ScreenTail.LineIter = t.GetTail()
	// }
	return t
}

// TODO написать присвоение id для курсора
// TODO на данном этапе при создании NewCursor просто всегда будем 0 в него передавать, когда будем писать серверную часть нужно будет определиться с созданием Id
func NewCursor(Id int64, ScreenRow int32) *Cursor {
	c := new(Cursor)
	c.Row = 0
	c.Col = -1
	c.Id = Id
	c.ScreenRow = ScreenRow
	c.ScreenHead = &Border{RowNumber: 0}
	c.ScreenTail = &Border{RowNumber: 0}
	return c
}

// TODO сделать красиво
//сделать откат строки, и ячейки если у нас не вставилась буква
// строку откатываем, только если мы одновременно создаём строку и ячейку
// иначе откатываем только ячейку

func (t *Text) InsertCharAfter(cursorId int64, value rune) error {
	cur := t.Cursors[cursorId]
	println("insert char after", cur.Row, cur.Col)

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
			println("row & col", cur.Row, cur.Col)
			err := t.InsertAfter(line, cur.LineIter)

			if err != nil {
				return err
			}

			cur.LineIter.GetValue().CopyTo(cur.LineIter.GetNext().GetValue())
			cur.LineIter.GetValue().Clear()

			cur.LineIter = cur.LineIter.GetNext()
			cur.CharIter = nil

			cur.ScrollDown()
			println("head & tail", cur.ScreenHead.RowNumber, cur.ScreenTail.RowNumber)

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

		cur.ScrollDown()

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
	// println("Next:", cur.LineIter.GetNext()," = NODE:", node)
	cur.LineIter = node
	// println("Head: ", cur.LineIter.GetValue().GetHead())
	cur.CharIter = nil

	cur.ScrollDown()

	cur.Row++
	cur.Col = -1
	return nil
}

// TODO cur.CharIter == nil не значит, что строка пустая || Нужно проверить везде в коде это
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

		cur.ScrollUp()

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
		charIter := iter.GetValue().GetHead()
		for charIter != nil {
			str += string(charIter.GetValue())
			charIter = charIter.GetNext()
		}
		str += "\n"
		iter = iter.GetNext()
	}
	//println("Original str: ", str)
	return str
}

// TODO think about width
func (t *Text) GetScreenString(cursorId int32) string {
	cur := t.Cursors[cursorId]
	println(cur.ScreenHead.LineIter, cur.ScreenTail.LineIter)
	if cur.ScreenHead.LineIter == nil || cur.ScreenTail.LineIter == nil {
		println("ScreenHead is nil")
		return ""
	}
	ptr1 := cur.ScreenHead.LineIter
	println(cur.ScreenHead, cur.ScreenTail)
	res := ""
	for ptr1 != nil && ptr1 != cur.ScreenTail.LineIter.GetNext() {
		ptr2 := ptr1.GetValue().GetHead()
		for ptr2 != nil {
			res += string(ptr2.GetValue())
			ptr2 = ptr2.GetNext()
		}
		res += "\n"
		ptr1 = ptr1.GetNext()
	}
	println("res: ", res)
	return res
}

func (cur *Cursor) MoveLeft() {
	if cur.CharIter == nil {
		if cur.LineIter.GetPrev() == nil {
			return
		}

		cur.LineIter = cur.LineIter.GetPrev()
		cur.CharIter = cur.LineIter.GetValue().GetTail()
		cur.Row--
		if cur.ScreenHead.LineIter.GetPrev() == cur.LineIter {
			cur.ScrollUp()
		}
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
			if cur.ScreenTail.LineIter.GetNext() == cur.LineIter {
				cur.ScrollDown()
			}
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
		if cur.ScreenTail.LineIter.GetNext() == cur.LineIter {
			cur.ScrollDown()
		}
		cur.Col = -1
		return
	}

	cur.CharIter = cur.CharIter.GetNext()
	cur.Col++
}

func (cur *Cursor) MoveUp() {
	if cur.LineIter.GetPrev() == nil {
		return
	}

	index := cur.LineIter.GetValue().Index(cur.CharIter)
	cur.CharIter = cur.LineIter.GetPrev().GetValue().GetNodeByIndex(index)
	cur.LineIter = cur.LineIter.GetPrev()
	cur.Row--
	if cur.ScreenHead.LineIter.GetPrev() == cur.LineIter {
		cur.ScrollUp()
	}
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
	if cur.ScreenTail.LineIter.GetNext() == cur.LineIter {
		cur.ScrollDown()
	}
	cur.Col = cur.LineIter.GetValue().Index(cur.CharIter)
}

func (cur *Cursor) ScrollUp() {
	if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 == cur.ScreenRow {
		println("here1")
		if cur.LineIter.GetNext() == cur.ScreenHead.LineIter {
			// Tail не должен быть равен nil, если Head не равен nil
			if cur.ScreenHead.LineIter.GetPrev() == nil || cur.ScreenTail.LineIter.GetPrev() == nil {
				return
			}
			cur.ScreenTail.LineIter = cur.ScreenTail.LineIter.GetPrev()
			cur.ScreenTail.RowNumber--
			cur.ScreenHead.LineIter = cur.ScreenHead.LineIter.GetPrev()
			cur.ScreenHead.RowNumber--
		} else {
			// Если мы попали в этот else, значит мы находимся на последней строке, и она поднимается наверх, а значит мы должны поменять ScreenTail
			cur.ScreenTail.LineIter = cur.ScreenTail.LineIter.GetPrev()
			cur.ScreenTail.RowNumber--
		}
	} else {
		println("here2")
		// cur.ScreenTail.LineIter = cur.ScreenTail.LineIter.GetPrev()
		cur.ScreenTail.RowNumber--
	}

	/* True ScrollUp
	   // Tail не должен быть равен nil, если Head не равен nil
	   if cur.ScreenHead.LineIter.GetPrev() == nil || cur.ScreenTail.LineIter.GetPrev() == nil {
	       return
	   }
	   cur.ScreenTail.LineIter = cur.ScreenTail.LineIter.GetPrev()
	   cur.ScreenTail.RowNumber--
	   cur.ScreenHead.LineIter = cur.ScreenHead.LineIter.GetPrev()
	   cur.ScreenHead.RowNumber--
	*/
}

func (cur *Cursor) ScrollDown() {
	println(cur.ScreenHead.RowNumber, cur.ScreenTail.RowNumber)
	if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 == cur.ScreenRow {
		//ЭТО НЕ КОСТЫЛЬ, это швабра, которая поддерживает потолок
		if cur.LineIter.GetPrev() == cur.ScreenTail.LineIter {
			// Head не должен быть равен nil, если Tail не равен nil
			if cur.ScreenHead.LineIter.GetNext() == nil || cur.ScreenTail.LineIter.GetNext() == nil {
				return
			}
			cur.ScreenTail.LineIter = cur.ScreenTail.LineIter.GetNext()
			cur.ScreenTail.RowNumber++
			cur.ScreenHead.LineIter = cur.ScreenHead.LineIter.GetNext()
			cur.ScreenHead.RowNumber++

		} else {
			// Если мы попали в этот else, то мы находимся в середине текста, нам нужно сделать ScrollDown, но нужно поменять только ScreenTail, Причём важно, что не нужно менять RowNumber
			cur.ScreenTail.LineIter = cur.ScreenTail.LineIter.GetPrev()
		}
	} else {
		println("aboba", cur.ScreenHead.RowNumber, cur.ScreenTail.RowNumber)
		cur.ScreenTail.LineIter = cur.ScreenTail.LineIter.GetNext()
		cur.ScreenTail.RowNumber++
	}
	/* Это тру scrollDown
	   // Head не должен быть равен nil, если Tail не равен nil
	   if cur.ScreenHead.LineIter.GetNext() == nil || cur.ScreenTail.LineIter.GetNext() == nil {
	       return
	   }
	   cur.ScreenTail.LineIter = cur.ScreenTail.LineIter.GetNext()
	   cur.ScreenTail.RowNumber++
	   cur.ScreenHead.LineIter = cur.ScreenHead.LineIter.GetNext()
	   cur.ScreenHead.RowNumber++
	*/
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
    res := ""hea
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
