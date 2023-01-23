package main

import (
	"bufio"
	"errors"
	"math"

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

type CharTexture struct {
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
    AllSuportedChars string = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,./()\\-+={}[]:;'"|?&*#<>`
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

// func RenderOneLine(PreRenderredChars map[rune]Cache,
//                 renderer *sdl.Renderer, line string,
//                 X int32, Y int32, FontSize int32,
//                 SpaceBetween int32) int32 {
//     var totatWidth int32 = 0
//     for _, el := range line {
//         texture := PreRenderredChars[rune(el)]
//         renderer.Copy(texture.Texture, nil, &sdl.Rect{X:totatWidth + X, Y:Y, W:texture.Width, H:FontSize})
//         totatWidth += texture.Width + SpaceBetween
//     }
//     return totatWidth
// }

type Cache struct {
    PreRenderredCharTextures    map[rune]CharTexture
    RectangleMatrix             *RectangleMatrix
}

func NewCache(supportedChars string, font *ttf.Font, windowWidth int32, windowHeight int32, renderer *sdl.Renderer) *Cache {
    cache := &Cache{}
    cache.PreRenderredCharTextures = make(map[rune]CharTexture)

    var (
        mn int32 = math.MaxInt32
        height int32
    )    

    for _, char := range supportedChars {
        fontSurface, _ := font.RenderUTF8Solid(string(char), sdl.Color{255, 0, 0, 255})
        texture, _ := renderer.CreateTextureFromSurface(fontSurface)
        cache.PreRenderredCharTextures[char] = CharTexture{texture, fontSurface.W}
        if fontSurface.W < mn {
            mn = fontSurface.W
        }
        height = fontSurface.H
    }

    var (
        rows = windowHeight / mn
        columns = windowWidth / height
    )

    cache.RectangleMatrix = NewRectangleMatrix(rows, columns)

    return cache 
}

type RectangleMatrix struct {
    RectangleMatrix [][]*sdl.Rect
    Rows            int32
    Columns         int32
}

func NewRectangleMatrix(rows, columns int32) *RectangleMatrix {
    rectangleMatrix := &RectangleMatrix{
        Rows: rows,
        Columns: columns,
    }
    rectangleMatrix.RectangleMatrix = make([][]*sdl.Rect, rows)
    for i := int32(0); i < rows; i++ {
        for j := int32(0); j < columns; j++ {
            rectangleMatrix.RectangleMatrix[i] = append(rectangleMatrix.RectangleMatrix[i], &sdl.Rect{})
        }
    }
    return rectangleMatrix
}


func main() {
    var (
        ScreenHeight int32 =      896
        ScreenWidth  int32 =      1200
        FontSize     int   =      52
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


    renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

    if err != nil {
        panic(err)
    }

    cache := NewCache(AllSuportedChars, font, ScreenWidth, ScreenHeight, renderer)

    text := "hello, world!" 
     
	running := true
    for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
            switch t := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
            case *sdl.TextInputEvent:
                pressedKey := t.GetText()
                text = text[:1] + pressedKey + text[2:]
                renderer.Clear()
                
                var (
                    X int32 = SpaceBetween
                    Y int32 = 0
                )

                for i, c := range text {
                    cache.RectangleMatrix.RectangleMatrix[0][i].H = int32(FontSize)
                    cache.RectangleMatrix.RectangleMatrix[0][i].W = cache.PreRenderredCharTextures[rune(c)].Width
                    cache.RectangleMatrix.RectangleMatrix[0][i].X = X
                    cache.RectangleMatrix.RectangleMatrix[0][i].Y = Y 
                    renderer.Copy(cache.PreRenderredCharTextures[rune(c)].Texture, nil, cache.RectangleMatrix.RectangleMatrix[0][i])
                    X += cache.PreRenderredCharTextures[rune(c)].Width + SpaceBetween
                }
                renderer.Present()
                break
            case *sdl.KeyboardEvent:
                // if len(textData) > 0 && t.Keysym.Scancode == 42 && t.State == sdl.PRESSED {
                //     textData = textData[:len(textData) - 1]
                // }
                //println(t.Keysym.Scancode)
                break
		    }
            //println(event)
        }
        sdl.Delay(50)
	}
    font.Close()
}

