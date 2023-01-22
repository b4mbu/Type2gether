package main

import (
	"bufio"
	"errors"
	//"fmt"
	"io"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
    endl = '\n'
)

type Char struct {
    value  rune
    next  *Char
    prev  *Char
}

type Line struct {
    value *Char
    next  *Line
    prev  *Line
    size   uint8
}

type Cursor struct {
    linePos *Line
    charPos *Char
    head    *Line
    id      int64
    // Xcoord
    // Tcoord
}

type Text struct {
    head  *Line
    size  uint32
    cursors []*Cursor
}

type Cache struct {
    Texture *sdl.Texture
    Width   int32
}


func CreateChar(value rune) *Char {
    char := new(Char)
    char.value = value

    return char
}

func CreateLine() *Line {
    line := new(Line)
    return line
}

func CreateCursor() *Cursor {
    c := new(Cursor)
    return c
}

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


var (
    AllSuportedChars string = "qwertyuiop[]asdfghjkl;zxcvbnm,./1234567890-=+"
)

func InitGUI() {
    if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

    if err := ttf.Init(); err != nil {
        panic(err)
    }
}

func EndGUI() {
    sdl.Quit()
    ttf.Quit()
}

func RenderOneLine(PreRenderredChars map[rune]Cache,
                renderer *sdl.Renderer, line string,
                X int32, Y int32, FontSize int32,
                SpaceBetween int32) int32 {
    var totatWidth int32 = 0
    for _, el := range line {
        texture := PreRenderredChars[rune(el)]
        renderer.Copy(texture.Texture, nil, &sdl.Rect{X:totatWidth + X, Y:Y, W:texture.Width, H:FontSize})
        totatWidth += texture.Width + SpaceBetween
    }
    return totatWidth
}

func main() {
    PreRenderredChars := make(map[rune]Cache)
    var (
        ScreenHeight int32 =      896
        ScreenWidth  int32 =      1200
        FontSize           =      52
        SpaceBetween int32 =      20
    )
    InitGUI()
    defer EndGUI()

    window, err := sdl.CreateWindow("Type2gether", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		 ScreenWidth,ScreenHeight, sdl.WINDOW_SHOWN)

	if err != nil {
		panic(err)
	}

	defer window.Destroy()

    font, err := ttf.OpenFont("nice.ttf", FontSize)
    if err != nil {
        panic(err)
    }


    fontSurface, _ := font.RenderUTF8Blended("hello, world", sdl.Color{255, 0, 0, 100})
    renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

    if err != nil {
        panic(err)
    }
// start position of the text
    var X int32 = 300
    var Y int32 = 120
// fill cache with textures
    for _, el := range AllSuportedChars {
        fontSurface, _ := font.RenderUTF8Blended(string(el), sdl.Color{255, 0, 0, 255})
        texture, _ := renderer.CreateTextureFromSurface(fontSurface)
        PreRenderredChars[el] = Cache{texture, fontSurface.W}
    }

    var textData string = "hello, world"

// speed of moving =)
    var (
        dx int32 = 5
        dy int32 = 8
    )

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
            switch t := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
            case *sdl.TextInputEvent:
                if len(textData) < 25 {
                    textData += t.GetText()
                }
                break
            case *sdl.KeyboardEvent:
                if len(textData) > 0 && t.Keysym.Scancode == 42 && t.State == sdl.PRESSED {
                    textData = textData[:len(textData) - 1]
                }
                //println(t.Keysym.Scancode)
                break
		    }
            //println(event)
        }

        Y += dy
        X += dx
        renderer.Clear()
        renderer.FillRect(nil)
        totatWidth := RenderOneLine(PreRenderredChars, renderer, textData,X,Y,int32(FontSize),SpaceBetween)
        // movement =)
        if X + dx <= 0 {
            dx *= -1
            X += 1
        }
        if X + totatWidth + dx >= ScreenWidth {
            dx *= -1
            X -= 1
        }

        if Y <= 0{
            dy *= -1
            Y += 1
        }
        if Y + int32(FontSize)  >= ScreenHeight {
            dy *= -1
            Y -= 1
        }

        renderer.Present()
        sdl.Delay(50)
	}
    font.Close()
    fontSurface.Free()
}

