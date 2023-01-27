package main

import (
	"math"
    "errors"
    "Type2gether/textmanager"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

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

// TODO add cursor.Color
func (e *Engine) RenderCursor() {
    var height int32 = e.font.GetSize()
    e.renderer.SetDrawColor(255, 255, 255, 255)

    for i := 0; i < len(e.text.Cursors); i++ {
        cur := e.text.Cursors[i]

        col := cur.Col
        row := cur.Row
        var padding int32
        if col == -1 {
            col = 0
            padding = 0
        } else {
            padding = e.GetRectFromMatrix(row, col).W
        }
        e.renderer.FillRect(&sdl.Rect{X: e.GetRectFromMatrix(row, col).X + padding, Y:e.GetRectFromMatrix(row, col).Y, W: 5, H: height})
    }

    e.renderer.SetDrawColor(0, 0, 0, 255)
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
        //cursor: &Cursor{0, -1},
    }
    // IMPORTANT FIX
    cur := textmanager.NewCursor(0)
    cur.LineIter = engine.text.GetHead()
    cur.CharIter = nil
    engine.text.Cursors = append(engine.text.Cursors, cur)

    err = engine.SetFont(fontFilename, fontSize, fontSpaceBetween, fontColor)

    if err != nil {
        return nil, err
    }

    err = engine.SetCache(supportedChars)

    if err != nil {
        return nil, err
    }

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
                // TODO FIX BUGGGGGG LEN(CURRENT LINE) < ....Columns
                break
            case *sdl.KeyboardEvent:
                if t.State != sdl.PRESSED {
                    break
                }

                if t.Keysym.Scancode == sdl.SCANCODE_BACKSPACE {
                    e.EraseChar(DEBUG_CUR_ID)
                } else if t.Keysym.Scancode == sdl.SCANCODE_RETURN {
                    e.InsertChar('\n', DEBUG_CUR_ID)
                } else if t.Keysym.Scancode == sdl.SCANCODE_TAB {
                    // TODO 4 -> SPACE IN ONE TAB
                    e.InsertChar('\t', DEBUG_CUR_ID)
                }

                e.renderText()
                break
		    }
        }
        sdl.Delay(50)
	}
}

// TODO add flag "Back space" or "Delete" and cursor ID
func (e *Engine) EraseChar (cursorId int64) {
/*
    cur := e.text.Cursors[cursorId]

    if  cur.LineIter.GetPrev() != nil {
        println("Line size() = ", cur.LineIter.GetValue().Length())
        
        return
    }
*/
    err := e.text.RemoveCharBefore(cursorId)
    if err != nil {
        println(err)
    }
}

func (e *Engine) InsertChar (value rune, cursorId int64) {
    cur := e.text.Cursors[cursorId]
    //res := e.text.InsertChar(value)
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

    for _, c := range e.text.GetString() {
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
    //println(e.cursor.col, e.cursor.row)
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
        FontColor    sdl.Color = sdl.Color{R: 255, G: 0, B: 0, A: 255}
        WindowTitle  string    = "Type2gether"
    )

    GUIStart()
    defer GUIStop()

    engine, err := NewEngine(ScreenWidth, ScreenHeight, FontFilename, FontSize, SpaceBetween, FontColor, WindowTitle, AllSupportedChars)

    if err != nil {
        panic(err)
    }

    //engine.SetText("")
    engine.Loop()
}

