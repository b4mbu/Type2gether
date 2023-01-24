package main

import (
	"math"
    "errors"
    //    "Type2gether/textmanager"
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

func NewRectangleMatrix(rows, columns int32, fontSize int) *RectangleMatrix {
    rectangleMatrix := &RectangleMatrix{
        Rows: rows,
        Columns: columns,
    }
    rectangleMatrix.RectangleMatrix = make([][]*sdl.Rect, rows)
    for i := int32(0); i < rows; i++ {
        for j := int32(0); j < columns; j++ {
            rectangleMatrix.RectangleMatrix[i] = append(rectangleMatrix.RectangleMatrix[i], &sdl.Rect{Y: i * int32(fontSize)})
        }
    }
    return rectangleMatrix
}

type Font struct {
    filename     string
    size         int
    color        sdl.Color
    spaceBetween int32
    ttfFont      *ttf.Font
}

func (f *Font) GetFilename() string {
    return f.filename
}

func (f *Font) GetSize() int {
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
    text        string
    cursor      *Cursor
}

type Cursor struct {
    row int32
    col int32
}

func (e *Engine) RenderCursor() {
    var height int32 = int32(e.font.GetSize())
    e.renderer.SetDrawColor(255, 255, 255, 255)
    println(e.cache.RectangleMatrix.RectangleMatrix[e.cursor.row][e.cursor.col].X)
    e.renderer.FillRect(&sdl.Rect{X: e.cache.RectangleMatrix.RectangleMatrix[e.cursor.row][e.cursor.col].X + e.cache.RectangleMatrix.RectangleMatrix[e.cursor.row][e.cursor.col].W , Y: e.cache.RectangleMatrix.RectangleMatrix[e.cursor.row][e.cursor.col].Y, W: 5, H: height})
    e.renderer.SetDrawColor(0, 0, 0, 255)
}

func (e *Engine) SetText(str string) {
    e.text = str
}

func NewEngine(windowWidth, windowHeight int32,
               fontFilename string,
               fontSize int,
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
        cursor: &Cursor{0, 0},
    }

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
}


func (e *Engine) SetFont(filename string, size int, spaceBetween int32, color sdl.Color) error {
    ttfFont, err := ttf.OpenFont(filename, size)

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

    cache.RectangleMatrix = NewRectangleMatrix(rows, columns, e.font.GetSize())
    e.cache = cache

    return nil
}

func (e *Engine) Loop() {
	running := true
    e.renderText()
    var fl int32 = 0
    for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
            switch t := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
            case *sdl.TextInputEvent:
                pressedKey := t.GetText()
                // TODO FIX BUGGGGGG LEN(CURRENT LINE) < ....Columns
                if len(e.text) < int(e.cache.RectangleMatrix.Columns) {
                    e.text += pressedKey
                    if fl == 0 {
                        fl = 1
                        e.renderText()
                    }else{
                        e.cursor.col += 1
                        e.renderText()
                    }
                    //e.cursor.col += 1
                    //e.renderText()
                }
                break
            case *sdl.KeyboardEvent:
                if len(e.text) > 0 && t.Keysym.Scancode == sdl.SCANCODE_BACKSPACE && t.State == sdl.PRESSED {
                    e.text = e.text[:len(e.text) - 1]
                    e.cursor.col -= 1
                   if e.cursor.col == -1 {
                       /* if e.cursor.row != 0 {
                            e.cursor.row -= 1
                            e.cursor.col = 0
                        }*/
                        e.cursor.col = 0
                    }
                    e.renderText()
                } else if t.Keysym.Scancode == sdl.SCANCODE_RETURN && t.State == sdl.PRESSED {
                    e.text = e.text + "\n"
                    //e.renderText()
                    e.cursor.col = 0
                    fl = 0
                    e.cursor.row += 1
                    e.renderText()
                }
                break
		    }
        }
        sdl.Delay(50)
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

    for _, c := range e.text {
        e.cache.RectangleMatrix.RectangleMatrix[row][col].H = int32(e.font.GetSize())
        e.cache.RectangleMatrix.RectangleMatrix[row][col].W = e.cache.PreRenderredCharTextures[rune(c)].Width
        e.cache.RectangleMatrix.RectangleMatrix[row][col].X = X
        e.cache.RectangleMatrix.RectangleMatrix[row][col].Y = Y
        e.renderer.Copy(e.cache.PreRenderredCharTextures[rune(c)].Texture, nil, e.cache.RectangleMatrix.RectangleMatrix[row][col])
        X += e.cache.PreRenderredCharTextures[rune(c)].Width + e.font.GetSpaceBetween()
        col++
        if c == '\n'  {
            Y += int32(e.font.GetSize())
            X = e.font.GetSpaceBetween()
            row++
            col = 0
        }
    }
    println(e.cursor.col, e.cursor.row)
    e.RenderCursor()
    e.renderer.Present()

}


func main() {
    var (
        ScreenHeight int32     = 896
        ScreenWidth  int32     = 1200
        FontSize     int       = 52
        SpaceBetween int32     = 10
        FontFilename string    = "nice.ttf"
        FontColor    sdl.Color = sdl.Color{255, 0, 0, 255}
        WindowTitle  string    = "Type2gether"
    )

    GUIStart()
    defer GUIStop()

    engine, err := NewEngine(ScreenWidth, ScreenHeight, FontFilename, FontSize, SpaceBetween, FontColor, WindowTitle, AllSupportedChars)

    if err != nil {
        panic(err)
    }
    
    engine.SetText("")
    engine.Loop()
}

