package textmanager

// 																																        сделать красиво =)

import (
	"client/list"
	"errors"
)

type Cursor struct {
	/*
	   	Cursor указывает на текущий элемент, а не на злемент за ним
	      	Line теперь представляем не как наследуемую структуру от list.List[rune], а как list.List[rune] сам по себе.
	      	Чтобы создать новую строку, надо использовать line := new(list.List[rune])
	*/
	LineIter *list.Node[*list.List[rune]]
	CharIter *list.Node[rune]
	Id       int64
	Row      int32
	Col      int32
	Color    uint32

	ScreenLeft int32

	ScreenHead *Border
	ScreenTail *Border
	ScreenRow  int32
	ScreenCol  int32
}

type Border struct {
	/* Эта структура нужна для скрола
	   Эта структура нужна для отрисовки нумерации строк */
	LineIter  *list.Node[*list.List[rune]]
	RowNumber int32
}

func (border *Border) Up() error {
	if border.LineIter == nil {
		return errors.New("lineIter is nil")
	}

	if border.LineIter.GetPrev() == nil {
		return errors.New("the first line")
	}

	border.LineIter = border.LineIter.GetPrev()
	border.RowNumber--
	return nil
}

func (border *Border) Down() error {
	if border.LineIter == nil {
		return errors.New("lineIter is nil")
	}

	if border.LineIter.GetNext() == nil {
		return errors.New("last line")
	}

	border.LineIter = border.LineIter.GetNext()
	border.RowNumber++
	return nil
}

func (cur *Cursor) SetPosition(lineIter *list.Node[*list.List[rune]], charIter *list.Node[rune], row, col int32) {
	/* Be careful! CharIter is not being checked whether it exists in lineIter */
	cur.LineIter = lineIter
	cur.CharIter = charIter
	cur.Row = row
	cur.Col = col
}

func (cur *Cursor) ScrollUp() {
	if err := cur.ScreenHead.Up(); err == nil {
		cur.ScreenTail.Up()
	}
}

func (cur *Cursor) ScrollDown() {
	if err := cur.ScreenTail.Down(); err == nil {
		cur.ScreenHead.Down()
	}
}

type Text struct {
	list.List[*list.List[rune]]
	Cursors []*Cursor
}

func NewText() *Text {
	t := new(Text)
	t.PushBack(new(list.List[rune]))
	return t
}

func NewCursor(Id int64, ScreenRow, ScreenCol int32) *Cursor {
	c := new(Cursor)
	c.Row = 0
	c.Col = -1
	c.Id = Id
	c.ScreenRow = ScreenRow
	c.ScreenCol = ScreenCol
	c.ScreenLeft = 0
	c.ScreenHead = &Border{RowNumber: 0}
	c.ScreenTail = &Border{RowNumber: 0}
	return c
}

func (t *Text) SetCursorStartPosition(cursorId int64) {
	cur := t.Cursors[cursorId]
	cur.SetPosition(t.GetHead(), nil, 0, -1)

	cur.ScreenLeft = 0
	cur.ScreenHead = &Border{LineIter: t.GetHead(), RowNumber: 0} // Егор сказал, что выстрелит, но куда... (создаем новый объект, а не меняем старый)

	cur.ScreenTail = &Border{LineIter: t.GetHead(), RowNumber: 0}
	for i := 1; i < int(t.Cursors[cursorId].ScreenRow); i++ {
		if err := cur.ScreenTail.Down(); err != nil {
			return
		}
	}
}

func (t *Text) InsertCharBefore(cursorId int64, value rune) error {
	err := t.InsertCharAfter(cursorId, value)
	if err != nil {
		return err
	}
	t.Cursors[cursorId].MoveLeft()
	return nil
}

func (t *Text) InsertCharAfter(cursorId int64, value rune) error {
	cur := t.Cursors[cursorId]

	if cur.CharIter == nil {
		err := cur.LineIter.GetValue().PushFront(value)

		if err != nil {
			return err
		}

		cur.SetPosition(cur.LineIter, cur.LineIter.GetValue().GetHead(), cur.Row, 0)
		return nil
	}

	err := cur.LineIter.GetValue().InsertAfter(value, cur.CharIter)

	if err != nil {
		return err
	}

	cur.SetPosition(cur.LineIter, cur.CharIter.GetNext(), cur.Row, cur.Col+1)
	return nil
}

// TODO validation
func (t *Text) Paste(data string, cursorId int64) error {
	var err error
	for _, c := range data {
		if c == '\n' {
			err = t.InsertLineAfter(cursorId)
		} else {
			err = t.InsertCharAfter(cursorId, c)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Text) InsertLineAfter(cursorId int64) error {
	cur := t.Cursors[cursorId]

	if cur.CharIter == nil {
		line := new(list.List[rune])
		if cur.LineIter.GetValue().GetTail() != nil {
			return t.InsertLineAfterNonEmptyLineFromBeginning(line, cur)
		}
		return t.InsertLineAfterEmptyLine(line, cur)
	}
	return t.InsertLineAfterNonEmptyLine(cur)
}

func (t *Text) InsertLineAfterNonEmptyLineFromBeginning(line *list.List[rune], cur *Cursor) error {
	err := t.InsertAfter(line, cur.LineIter)

	if err != nil {
		return err
	}

	t.InsertLineAfterScroll(cur)

	cur.LineIter.GetValue().CopyTo(cur.LineIter.GetNext().GetValue())
	cur.LineIter.GetValue().Clear()

	cur.SetPosition(cur.LineIter.GetNext(), nil, cur.Row+1, -1)
	return nil
}

func (t *Text) InsertLineAfterEmptyLine(line *list.List[rune], cur *Cursor) error {
	err := t.InsertAfter(line, cur.LineIter)

	if err != nil {
		return err
	}

	t.InsertLineAfterScroll(cur)

	cur.SetPosition(cur.LineIter.GetNext(), nil, cur.Row+1, -1)
	return nil
}

func (t *Text) InsertLineAfterNonEmptyLine(cur *Cursor) error {
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

	t.InsertLineAfterScroll(cur)

	cur.SetPosition(node, nil, cur.Row+1, -1)
	return nil
}

func (t *Text) InsertLineAfterScroll(cur *Cursor) {
	if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 >= cur.ScreenRow {
		if cur.LineIter == cur.ScreenTail.LineIter {
			cur.ScrollDown()
		} else {
			cur.ScreenTail.Up()
			cur.ScreenTail.RowNumber++
		}
	} else {
		if cur.LineIter == cur.ScreenTail.LineIter {
			cur.ScreenTail.Down()
		} else {
			cur.ScreenTail.RowNumber++
		}
	}
}

// TODO cur.CharIter == nil не значит, что строка пустая || Нужно проверить везде в коде это
func (t *Text) RemoveCharBefore(cursorId int64) error {
	cur := t.Cursors[cursorId]

	if cur.LineIter == nil {
		//printtln("Press 'F', пацан к успеху шёл, Какой успех, раздался смех")
		return nil
	}

	if cur.CharIter == nil {
		if cur.LineIter.GetPrev() == nil {
			return nil
		}

		if cur.ScreenTail.LineIter == cur.ScreenHead.LineIter {
			cur.ScrollUp()
		} else if cur.ScreenTail.LineIter == cur.LineIter {
			if cur.LineIter.GetNext() == nil {
				cur.ScreenTail.Up()
			} else {
				if err := cur.ScreenTail.Down(); err == nil {
					cur.ScreenTail.RowNumber--
				}
			}
		} else if cur.ScreenHead.LineIter == cur.LineIter {
			if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 >= cur.ScreenRow {
				if err := cur.ScreenHead.Up(); err == nil {
					cur.ScreenTail.RowNumber--
				}
			} else {
				if err := cur.ScreenHead.Up(); err == nil {
					cur.ScreenTail.RowNumber--
				}
			}
		} else {
			if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 >= cur.ScreenRow {
				err := cur.ScreenTail.Down()
				if err != nil {
				}
				cur.ScreenTail.RowNumber--
			} else {
				cur.ScreenTail.RowNumber--
			}
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

func (t *Text) RemoveCharAfter(cursorId int64) error {
	cur := t.Cursors[cursorId]
	oldCharIter := cur.CharIter
	oldLineIter := cur.LineIter
	cur.MoveRight()
	if oldCharIter != cur.CharIter || oldLineIter != cur.LineIter {
		err := t.RemoveCharBefore(cursorId)
		return err
	}
	return nil
}

// курсор не двигается!!
func (t *Text) MergeLines(cursorId int64) error {
	cur := t.Cursors[cursorId]
	if cur.LineIter.GetPrev() == nil {
		return errors.New("there is no previous line")
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
		iter = iter.GetNext()
		if iter != nil {
			str += "\n"
		}
	}
	////printtln("Original str: ", str)
	return str
}

// TODO think about width
func (t *Text) GetScreenString(cursorId int32) string {
	cur := t.Cursors[cursorId]
	//printtln(cur.ScreenHead.LineIter, cur.ScreenTail.LineIter)
	if cur.ScreenHead.LineIter == nil || cur.ScreenTail.LineIter == nil {
		//printtln("ScreenHead is nil")
		return ""
	}
	ptr1 := cur.ScreenHead.LineIter
	//printtln(cur.ScreenHead, cur.ScreenTail)
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
	//printtln("res: ", res)
	return res
}

func (cur *Cursor) MoveLeft() {
	if cur.CharIter == nil {
		if cur.LineIter.GetPrev() == nil {
			return
		}

		if cur.ScreenHead.LineIter == cur.LineIter {
			if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 >= cur.ScreenRow {
				cur.ScrollUp()
			} else {
				cur.ScreenHead.Up()
			}
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

			if cur.ScreenTail.LineIter == cur.LineIter {
				if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 >= cur.ScreenRow {
					cur.ScrollDown()
				}
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

		if cur.ScreenTail.LineIter == cur.LineIter {
			if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 >= cur.ScreenRow {
				cur.ScrollDown()
			}
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

func (cur *Cursor) MoveUp() {
	if cur.LineIter.GetPrev() == nil {
		return
	}

	if cur.ScreenHead.LineIter == cur.LineIter {
		if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 >= cur.ScreenRow {
			cur.ScrollUp()
		} else {
			cur.ScreenHead.Up()
		}
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

	if cur.ScreenTail.LineIter == cur.LineIter {
		if cur.ScreenTail.RowNumber-cur.ScreenHead.RowNumber+1 >= cur.ScreenRow {
			cur.ScrollDown()
		}
	}

	index := cur.LineIter.GetValue().Index(cur.CharIter)
	cur.CharIter = cur.LineIter.GetNext().GetValue().GetNodeByIndex(index)
	cur.LineIter = cur.LineIter.GetNext()
	cur.Row++
	cur.Col = cur.LineIter.GetValue().Index(cur.CharIter)
}

func (cur *Cursor) MoveHome() {
	cur.CharIter = nil
	cur.Col = -1
}

func (cur *Cursor) MoveEnd() {
	cur.CharIter = cur.LineIter.GetValue().GetTail()
	cur.Col = cur.LineIter.GetValue().Length() - 1
}
