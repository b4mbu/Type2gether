package textmanager

import (
	"Type2gether/list"
	"errors"
)

const (
	endl = '\n'
    PURPLE  = 0xDC42DDFF
    RED     = 0x51D718FF
    RED2    = 0xFF8300FF
    BLUE    = 0x8A6CD5FF
    CIAN    = 0xF5A58DFF
    PINK    = 0x0893FCFF
    FNTCLR  = 0xFFFFFFFF
    LNNMBC  = 0xbd93f9FF
)

var COLORS = []uint32{PURPLE, RED, BLUE, CIAN, PINK, FNTCLR, LNNMBC, RED2}

var (
	TokensColor = map[string]uint32{
		"for": PURPLE,
		"while": RED,
		"if": BLUE,
		"else": BLUE,
		"return": RED2,
		"int": CIAN,
		"long": CIAN,
		"aboba": PINK,
	}
	MaxTokenLength int
)

type Char struct {
	Value rune
	Color uint32
}

type Cursor struct {
	LineIter *list.Node[*list.List[Char]]
	CharIter *list.Node[Char]
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

//Эта структура нужна для скрола
//Эта структура нужна для отрисовки нумерации строк
type Border struct {
	LineIter  *list.Node[*list.List[Char]]
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

/*
type Line struct {
    Value       *list.List[rune]
    IsNatural   bool
    width       int32
}
*/

type Text struct {
	list.List[*list.List[Char]]
	// скорее всего не нужен тк в List есть аттрибут lenght /  size      int64
	Cursors []*Cursor
}

func NewText() *Text {
	t := new(Text)
	t.PushBack(new(list.List[Char]))
	return t
}

// TODO написать присвоение id для курсора
// TODO на данном этапе при создании NewCursor просто всегда будем 0 в него передавать, когда будем писать серверную часть нужно будет определиться с созданием Id
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
	cur.Row = 0
	cur.LineIter = t.GetHead()

	cur.Col = -1
	cur.CharIter = nil

	cur.ScreenLeft = 0
	cur.ScreenHead = &Border{LineIter: t.GetHead(), RowNumber: 0} // Егор сказал, что выстрелит, но куда... (создаем новый объект, а не меняем старый)

	cur.ScreenTail = &Border{LineIter: t.GetHead(), RowNumber: 0}
	for i := 1; i < int(t.Cursors[cursorId].ScreenRow); i++ {
		if err := cur.ScreenTail.Down(); err != nil {
			return
		}
	}
}

// TODO сделать красиво =)
//сделать откат строки, и ячейки если у нас не вставилась буква
// строку откатываем, только если мы одновременно создаём строку и ячейку
// иначе откатываем только ячейку

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
		err := cur.LineIter.GetValue().PushFront(Char{Value: value, Color: 0x000000FF})

		if err != nil {
			//printtln(err)
			return err
		}

		cur.CharIter = cur.LineIter.GetValue().GetHead()
		cur.Col = 0
		return nil
	}

	err := cur.LineIter.GetValue().InsertAfter(Char{Value: value, Color: 0x000000FF}, cur.CharIter)

	if err != nil {
		return err
	}

	cur.CharIter = cur.CharIter.GetNext()
	cur.Col++
	return nil
}

// TODO validation
func (t *Text) Paste(data string, cursorId int64) error {
	//cur := t.Cursors[cursorId]
	var err error
	//println("data: ", data)
	// TODO замена на before и перевернуть цикл
	for _, e := range data {
		if e == '\n' {
			//println("\\n")
			err = t.InsertLineAfter(cursorId)
		} else {
			err = t.InsertCharAfter(cursorId, e)
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
		line := new(list.List[Char])
		if cur.LineIter.GetValue().GetTail() != nil {
			// line isn't empty
			err := t.InsertAfter(line, cur.LineIter)

			if err != nil {
				return err
			}

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

			cur.LineIter.GetValue().CopyTo(cur.LineIter.GetNext().GetValue())
			cur.LineIter.GetValue().Clear()

			cur.LineIter = cur.LineIter.GetNext()
			cur.CharIter = nil

			// cur.ScrollDown()
			//printtln("head & tail", cur.ScreenHead.RowNumber, cur.ScreenTail.RowNumber)

			cur.Row++
			cur.Col = -1
			return nil
		}
		// line is empty
		err := t.InsertAfter(line, cur.LineIter)

		if err != nil {
			return err
		}

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

		cur.LineIter = cur.LineIter.GetNext()
		cur.CharIter = cur.LineIter.GetValue().GetHead()

		// cur.ScrollDown()

		cur.Row++
		cur.Col = -1
		return nil
	}

	oldLineIter := cur.LineIter
	lst, err := oldLineIter.GetValue().Split(cur.CharIter.GetNext())

	if err != nil {
		return err
	}

	node := &list.Node[*list.List[Char]]{}
	node.SetValue(lst)

	node.SetPrev(oldLineIter)
	node.SetNext(oldLineIter.GetNext())
	if oldLineIter.GetNext() != nil {
		oldLineIter.GetNext().SetPrev(node)
	} else {
		t.SetTail(node)
	}
	oldLineIter.SetNext(node)

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

	cur.LineIter = node
	cur.CharIter = nil

	// cur.ScrollDown()

	cur.Row++
	cur.Col = -1
	return nil
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
			str += string(charIter.GetValue().Value)
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
			res += string(ptr2.GetValue().Value)
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
