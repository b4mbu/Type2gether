package main

import (
	"math"
    "errors"
    "Type2gether/textmanager"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func hexToSdlColor(color uint32) sdl.Color {
    r, g, b, a := hexToRBGA(color)
    return sdl.Color{R: r, G: g, B: b, A: a}
}

func hexToRBGA(color uint32) (uint8, uint8, uint8, uint8) {
    r:= uint8((color>>(3*8)) & ((1<<8) - 1))
    g:= uint8((color>>(2*8)) & ((1<<8) - 1))
    b:= uint8((color>>(1*8)) & ((1<<8) - 1))
    a:= uint8((color>>(0*8)) & ((1<<8) - 1))
    return r, g, b, a
}


type CharTexture struct {
    Texture *sdl.Texture
    Width   int32
}

var (
    AllSupportedChars string = " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-+*/!#$~<>{}[]();,.|?:^%&@=_№`\\'\""
)

func GUIStart() {
    if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

    if err := ttf.Init(); err != nil {
        panic(err)
    }
}

func GUIStop() {
    sdl.Quit()
    ttf.Quit()
}

type Cache struct {
    PreRenderredCharTextures    map[rune]CharTexture
    RectangleMatrix             *RectangleMatrix
}

type RectangleMatrix struct {
    RectangleMatrix [][]*sdl.Rect
    Rows            int32
    Columns         int32
}

func NewRectangleMatrix(rows, columns, fontSize, SpaceBetween int32) *RectangleMatrix {
    rectangleMatrix := &RectangleMatrix{
        Rows: rows,
        Columns: columns,
    }
    rectangleMatrix.RectangleMatrix = make([][]*sdl.Rect, rows)
    for i := int32(0); i < rows; i++ {
        for j := int32(0); j < columns; j++ {
            rectangleMatrix.RectangleMatrix[i] = append(rectangleMatrix.RectangleMatrix[i], &sdl.Rect{X: SpaceBetween, Y: i * fontSize})
        }
    }
    return rectangleMatrix
}

type Font struct {
    filename     string
    size         int32
    color        sdl.Color
    spaceBetween int32
    ttfFont      *ttf.Font
}

func (f *Font) GetFilename() string {
    return f.filename
}

func (f *Font) GetSize() int32 {
    return f.size
}

func (f *Font) GetColor() sdl.Color {
    return f.color
}

func (f *Font) GetSpaceBetween() int32 {
    return f.spaceBetween
}

type Engine struct {
    cache       *Cache
    renderer    *sdl.Renderer
    window      *sdl.Window
    font        Font
    text        *textmanager.Text
}

func (e *Engine) RenderCursor() {
    var height int32 = e.font.GetSize()

    for i := 0; i < len(e.text.Cursors); i++ {
        cur := e.text.Cursors[i]
        e.renderer.SetDrawColor(hexToRBGA(cur.Color))
        col := cur.Col
        row := cur.Row
        var padding int32
/*        if row > 16 {
            cur.Row = 16
            row = 16
            println("AXTYNG!!!")
        }
*/
        if col == -1 {
            col = 0
            padding = 0
        } else {
            padding = e.GetRectFromMatrix(row, col).W
        }
        e.renderer.FillRect(&sdl.Rect{X: e.GetRectFromMatrix(row, col).X + padding, Y:e.GetRectFromMatrix(row, col).Y, W: 5, H: height})
    }

    e.renderer.SetDrawColor(hexToRBGA(0x000000FF))
}

func (e *Engine) GetRectFromMatrix(row, col int32) *sdl.Rect {
    return e.cache.RectangleMatrix.RectangleMatrix[row][col]
}
/*
func (e *Engine) SetText(str string) {
    e.text = str
}*/
// TODO add NEW cursor
func NewEngine(windowWidth, windowHeight int32,
               fontFilename string,
               fontSize int32,
               fontSpaceBetween int32,
               fontColor sdl.Color, windowTitle, supportedChars string) (*Engine, error) {

    window, err := sdl.CreateWindow(windowTitle, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, windowWidth, windowHeight, sdl.WINDOW_SHOWN)

    if err != nil {
        return nil, err
    }

    renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

    if err != nil {
        return nil, err
    }

    engine := &Engine{
        renderer: renderer,
        window: window,
        text : textmanager.NewText(),
    }

    err = engine.SetFont(fontFilename, fontSize, fontSpaceBetween, fontColor)

    if err != nil {
        return nil, err
    }

    err = engine.SetCache(supportedChars)

    if err != nil {
        return nil, err
    }

    // TODO server.createCursor(id) ??
    cur := textmanager.NewCursor(0, engine.cache.RectangleMatrix.Rows)
    cur.LineIter = engine.text.GetHead()
    cur.CharIter = nil
    cur.Color = 0x61A8DCFF
    cur.ScreenHead = engine.text.GetHead()
    cur.ScreenTail = engine.text.GetTail()
    engine.text.Cursors = append(engine.text.Cursors, cur)
    return engine, nil
}

func (e *Engine) Stop() {
    if e.window != nil {
        e.window.Destroy()
    }

    if e.font.ttfFont != nil {
        e.font.ttfFont.Close()
    }

    if e.cache != nil {
        for _, texture := range e.cache.PreRenderredCharTextures {
            if texture.Texture == nil {
                continue
            }
            texture.Texture.Destroy()
        }
    }
    // TODO think about dele all text ???
}


func (e *Engine) SetFont(filename string, size int32, spaceBetween int32, color sdl.Color) error {
    ttfFont, err := ttf.OpenFont(filename, int(size))

    if err != nil {
        return err
    }

    if e.font.ttfFont != nil {
        e.font.ttfFont.Close()
    }

    e.font = Font{
        filename: filename,
        size:     size,
        color:    color,
        ttfFont:  ttfFont,
        spaceBetween: spaceBetween,
    }

    return nil
}

func (e *Engine) SetCache(supportedChars string) error {
    if e.window == nil || e.font.ttfFont == nil || e.renderer == nil {
        return errors.New("field(s) not initialized")
    }

    cache := &Cache{}
    cache.PreRenderredCharTextures = make(map[rune]CharTexture)

    var (
        mn     int32 = math.MaxInt32
        height int32
    )

    for _, char := range supportedChars {
        fontSurface, _ := e.font.ttfFont.RenderUTF8Solid(string(char), e.font.GetColor())
        texture, _ := e.renderer.CreateTextureFromSurface(fontSurface)
        cache.PreRenderredCharTextures[char] = CharTexture{texture, fontSurface.W}
        if fontSurface.W < mn {
            mn = fontSurface.W
        }
        height = fontSurface.H
    }
    var (
        w, h = e.window.GetSize()
        rows    = h / height
        columns = w / (mn + e.font.GetSpaceBetween())
    )

    cache.RectangleMatrix = NewRectangleMatrix(rows, columns, e.font.GetSize(), e.font.GetSpaceBetween())
    e.cache = cache

    return nil
}

func (e *Engine) Loop() {
	running := true
    e.renderText()
    DEBUG_CUR_ID := int64(0)
    for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
            switch t := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
                break
            case *sdl.TextInputEvent:
                pressedKey := t.GetText()
                e.InsertChar(rune(pressedKey[0]), DEBUG_CUR_ID)
                e.renderText()
                break
            case *sdl.KeyboardEvent:
                // this branch active too when *sdl.TextInputEvent
                // TODO be careful!
                if t.State != sdl.PRESSED {
                    break
                }
                println("Scancode: ", t.Keysym.Scancode)

                switch t.Keysym.Scancode {
                case sdl.SCANCODE_BACKSPACE:
                    e.EraseChar(DEBUG_CUR_ID)

                case sdl.SCANCODE_RETURN:
                    e.InsertChar('\n', DEBUG_CUR_ID)

                case sdl.SCANCODE_TAB:
                    // TODO 4 -> SPACE_IN_ONE_TAB
                    e.InsertChar('\t', DEBUG_CUR_ID)

                case sdl.SCANCODE_LEFT:
                    e.MoveCursor("left", DEBUG_CUR_ID)

                case sdl.SCANCODE_RIGHT:
                    e.MoveCursor("right", DEBUG_CUR_ID)

                case sdl.SCANCODE_UP:
                    e.MoveCursor("up", DEBUG_CUR_ID)

                case sdl.SCANCODE_DOWN:
                    e.MoveCursor("down", DEBUG_CUR_ID)
                }
                e.renderText()
                break
		    }
        }
        sdl.Delay(50)
	}
}

func (e *Engine) MoveCursor (direction string, cursorId int64) {
    cur := e.text.Cursors[cursorId]
    if direction == "left" {
        cur.MoveLeft()
    } else if direction == "right" {
        cur.MoveRight()
    } else if direction == "up" {
        cur.MoveUp()
    } else if direction == "down" {
        cur.MoveDown()
    }
}

// TODO add flag "Back space" or "Delete"
func (e *Engine) EraseChar (cursorId int64) {
    err := e.text.RemoveCharBefore(cursorId)
    if err != nil {
        println(err)
    }
}

func (e *Engine) InsertChar (value rune, cursorId int64) {
    cur := e.text.Cursors[cursorId]
    if value == '\n' {
        err := e.text.InsertLineAfter(cursorId)
        if err != nil {
            println(err)
        }
    } else if value == '\t' {
        for i:= 0; i < 4 ; i++ {
            e.InsertChar(' ', cursorId)
        }

    } else {
        if cur.LineIter.GetValue().Length() + 1 < e.cache.RectangleMatrix.Columns {
            err := e.text.InsertCharAfter(cursorId, value)
            if err != nil {
                println(err)
            }
        }
    }
}

func (e *Engine) renderText() {
    e.renderer.Clear()

    var (
        X   int32 = e.font.GetSpaceBetween()
        Y   int32 = 0
        row int32 = 0
        col int32 = 0
    )

//  for _, c := range e.text.GetString() {
    for _, c := range e.text.GetScreenString(0) {
        e.GetRectFromMatrix(row, col).H = e.font.GetSize()
        e.GetRectFromMatrix(row, col).W = e.cache.PreRenderredCharTextures[rune(c)].Width
        e.GetRectFromMatrix(row, col).X = X
        e.GetRectFromMatrix(row, col).Y = Y
        e.renderer.Copy(e.cache.PreRenderredCharTextures[rune(c)].Texture, nil, e.GetRectFromMatrix(row, col))
        X += e.cache.PreRenderredCharTextures[rune(c)].Width + e.font.GetSpaceBetween()
        col++
        if c == '\n'  {
            Y += e.font.GetSize()
            X = e.font.GetSpaceBetween()
            row++
            col = 0
        }
    }
    e.RenderCursor()
    e.renderer.Present()

}


func main() {
    var (
        ScreenHeight int32     = 896
        ScreenWidth  int32     = 1200
        FontSize     int32     = 52
        SpaceBetween int32     = 10
        FontFilename string    = "nice.ttf"
        FontColor    uint32    = 0xFF0000FF
        WindowTitle  string    = "Type2gether"
    )

    GUIStart()
    defer GUIStop()

    engine, err := NewEngine(ScreenWidth, ScreenHeight, FontFilename, FontSize, SpaceBetween, hexToSdlColor(FontColor), WindowTitle, AllSupportedChars)

    if err != nil {
        panic(err)
    }

    engine.Loop()
}
