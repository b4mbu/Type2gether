package textmanager

// 																																        сделать красиво =)

import (
	"errors"
	"server/list"
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
}

func NewCursor(Id int64) *Cursor {
	c := new(Cursor)
	c.Row = 0
	c.Col = -1
	c.Id = Id
	return c
}
func (cur *Cursor) SetPosition(lineIter *list.Node[*list.List[rune]], charIter *list.Node[rune], row, col int32) {
	/* Be careful! CharIter is not being checked whether it exists in lineIter */
	cur.LineIter = lineIter
	cur.CharIter = charIter
	cur.Row = row
	cur.Col = col
}

func (cur *Cursor) MoveLeft() {
	if cur.CharIter == nil {
		if cur.LineIter.GetPrev() == nil {
			return
		}

		cur.SetPosition(cur.LineIter.GetPrev(), cur.LineIter.GetPrev().GetValue().GetTail(), cur.Row-1, cur.LineIter.GetPrev().GetValue().Length()-1)
		return
	}

	cur.SetPosition(cur.LineIter, cur.CharIter.GetPrev(), cur.Row, cur.Col-1)
}

func (cur *Cursor) MoveRight() {
	if cur.CharIter == nil {
		if cur.LineIter.GetValue().GetHead() == nil {
			if cur.LineIter.GetNext() == nil {
				return
			}

			cur.SetPosition(cur.LineIter.GetNext(), nil, cur.Row+1, -1)
			return
		}
		cur.SetPosition(cur.LineIter, cur.LineIter.GetValue().GetHead(), cur.Row, 0)
		return
	}

	if cur.CharIter == cur.LineIter.GetValue().GetTail() {
		if cur.LineIter.GetNext() == nil {
			return
		}

		cur.SetPosition(cur.LineIter.GetNext(), nil, cur.Row+1, -1)
		return
	}
	cur.SetPosition(cur.LineIter, cur.CharIter.GetNext(), cur.Row, cur.Col+1)
}

func (cur *Cursor) MoveUp() {
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

func (cur *Cursor) MoveHome() {
	cur.SetPosition(cur.LineIter, nil, cur.Row, -1)
}

func (cur *Cursor) MoveEnd() {
	cur.SetPosition(cur.LineIter, cur.LineIter.GetValue().GetTail(), cur.Row, cur.LineIter.GetValue().Length()-1)
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

func (t *Text) SetCursorStartPosition(cursorId int64) {
	cur := t.Cursors[cursorId]
	cur.SetPosition(t.GetHead(), nil, 0, -1)
}

func (t *Text) AddNewCursor(cursorId int64) {
	cursor := NewCursor(cursorId)
	t.Cursors = append(t.Cursors, cursor)
	t.SetCursorStartPosition(cursorId)
}

func (t *Text) RemoveCursor(cursorId int64) {
	if cursorId+1 >= int64(len(t.Cursors)) {
		t.Cursors = t.Cursors[:len(t.Cursors)-1]
		return
	}
	t.Cursors = append(t.Cursors[:cursorId], t.Cursors[cursorId+1:]...)
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

	if value == '\n' {
		return t.InsertLineAfter(cursorId)
	} else if value == '\t' {
		for i := 0; i < 4; i++ {
			t.InsertCharAfter(cursorId, ' ')
		}
	}

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
		err = t.InsertCharAfter(cursorId, c)

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

	cur.SetPosition(node, nil, cur.Row+1, -1)
	return nil
}

func (t *Text) RemoveCharBefore(cursorId int64) error {
	cur := t.Cursors[cursorId]

	if cur.CharIter == nil {
		if cur.LineIter.GetPrev() == nil {
			return nil
		}
		return t.RemoveCharBeforeFromBeginning(cur)
	}
	return t.RemovwCharBeforeFromMiddle(cur)
}

func (t *Text) RemoveCharBeforeFromBeginning(cur *Cursor) error {

	prev := cur.LineIter.GetPrev()
	oldTail := prev.GetValue().GetTail()
	oldLen := prev.GetValue().Length()
	err := t.MergeLines(cur.Id)
	if err != nil {
		return err
	}

	cur.SetPosition(prev, oldTail, cur.Row-1, oldLen-1)
	return nil
}

func (t *Text) RemovwCharBeforeFromMiddle(cur *Cursor) error {
	newCharIter := cur.CharIter.GetPrev()
	err := cur.LineIter.GetValue().Remove(cur.CharIter)

	if err != nil {
		return err
	}
	cur.SetPosition(cur.LineIter, newCharIter, cur.Row, cur.Col-1)
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

func (t *Text) MergeLines(cursorId int64) error {
	/* Be careful cursor does not move */
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

func (t *Text) GetFullText() string {
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
	return str
}
