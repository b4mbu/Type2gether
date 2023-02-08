package main

import (
	"Type2gether/textmanager"
	"errors"
	"math"
    "Type2gether/filemanager"
    "os"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func hexToSdlColor(color uint32) sdl.Color {
	r, g, b, a := hexToRBGA(color)
	return sdl.Color{R: r, G: g, B: b, A: a}
}

func hexToRBGA(color uint32) (uint8, uint8, uint8, uint8) {
	r := uint8((color >> (3 * 8)) & ((1 << 8) - 1))
	g := uint8((color >> (2 * 8)) & ((1 << 8) - 1))
	b := uint8((color >> (1 * 8)) & ((1 << 8) - 1))
	a := uint8((color >> (0 * 8)) & ((1 << 8) - 1))
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
	PreRenderredCharTextures map[rune]CharTexture
	PreRenderredNumsTextures map[rune]CharTexture
	RectangleMatrix          *RectangleMatrix
}

type RectangleMatrix struct {
	RectangleMatrix [][]*sdl.Rect
	Rows            int32
	Columns         int32
}

func NewRectangleMatrix(rows, columns, fontSize, SpaceBetween int32) *RectangleMatrix {
	rectangleMatrix := &RectangleMatrix{
		Rows:    rows,
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
    numsColor    sdl.Color
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

func (f* Font) GetNumColor() sdl.Color {
    return f.numsColor
}

func (f *Font) GetSpaceBetween() int32 {
	return f.spaceBetween
}


type Engine struct {
	cache    *Cache
	renderer *sdl.Renderer
	window   *sdl.Window
	font     Font
	text     *textmanager.Text
    filename string
}

func (e *Engine) RenderCursor(cursorId int64) {
	var (
        height int32 = e.font.GetSize()
        //delta  int32 = e.text.Cursors[cursorId].Col - e.cache.RectangleMatrix.Columns + 1 + 4
    )

    leftBorder := e.text.Cursors[cursorId].ScreenLeft
    //println("LeftBorder: ", leftBorder)

	for i := 0; i < len(e.text.Cursors); i++ {
		cur := e.text.Cursors[i]
		//println("Cur head: ", cur.ScreenHead.RowNumber, "Cur tail: ", cur.ScreenTail.RowNumber)
		e.renderer.SetDrawColor(hexToRBGA(cur.Color))
		col := cur.Col
// TODO отображать относительно Одного курсора, поэтому cur.ScreenHead.RowNumber не годится (относительно самого себя)
		row := cur.Row - cur.ScreenHead.RowNumber
		var padding int32

        // add this CHECK [not shure]
        if row < 0 || row >= e.cache.RectangleMatrix.Rows {
            continue
        }

		if col == -1 {
			col = 4 - leftBorder
			padding = 0
		} else {
            if col < leftBorder - 1 {
                continue
            }
            if col == leftBorder - 1 && leftBorder != 0 {
                //println("\t\t\tLEFTBORDER")
                col++
			    padding = 0
                col -= leftBorder
                col += 4
            } else {
                col -= leftBorder
                col += 4
			    padding = e.GetRectFromMatrix(row, col).W
            }
        }
		e.renderer.FillRect(&sdl.Rect{X: e.GetRectFromMatrix(row, col).X + padding, Y: e.GetRectFromMatrix(row, col).Y, W: 5, H: height})
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
	fontColor sdl.Color,
    numsColor sdl.Color, windowTitle, supportedChars string) (*Engine, error) {


    if len(os.Args[1:]) < 1 {
        return nil, errors.New("No file name. Nothing can be openned/saved")
    }

	window, err := sdl.CreateWindow(windowTitle, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, windowWidth, windowHeight, sdl.WINDOW_SHOWN)

//    window.SetFullscreen(1)

	if err != nil {
		return nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	if err != nil {
		return nil, err
	}

	engine := &Engine{
		renderer: renderer,
		window:   window,
		text:     textmanager.NewText(),
	}

	err = engine.SetFont(fontFilename, fontSize, fontSpaceBetween, fontColor, numsColor)

	if err != nil {
		return nil, err
	}

	err = engine.SetCache(supportedChars)

	if err != nil {
		return nil, err
	}

	// TODO server.createCursor(id) ??
	cur := textmanager.NewCursor(0, engine.cache.RectangleMatrix.Rows, engine.cache.RectangleMatrix.Columns)
	cur.LineIter = engine.text.GetHead()
	cur.CharIter = nil
	cur.Color = 0x61A8DCFF
	// TODO для server сделать нормальное присвоение Tail и Head
	cur.ScreenHead.LineIter = engine.text.GetHead()
	cur.ScreenTail.LineIter = engine.text.GetTail()
	engine.text.Cursors = append(engine.text.Cursors, cur)
    // TODO add regular expr!!!
    engine.filename = os.Args[1]
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

func (e *Engine) SetFont(filename string, size int32, spaceBetween int32, color sdl.Color, numsColor sdl.Color) error {
	ttfFont, err := ttf.OpenFont(filename, int(size))

	if err != nil {
		return err
	}

	if e.font.ttfFont != nil {
		e.font.ttfFont.Close()
	}

	e.font = Font{
		filename:     filename,
		size:         size,
		color:        color,
        numsColor:    numsColor,
		ttfFont:      ttfFont,
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
	cache.PreRenderredNumsTextures = make(map[rune]CharTexture)

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
		height = e.font.size
	}

    // cache all digits for rendering Numbers of Rows
    for _, char := range "0123456789" {
		numsSurface, _ := e.font.ttfFont.RenderUTF8Solid(string(char), e.font.GetNumColor())
		texture, _ := e.renderer.CreateTextureFromSurface(numsSurface)
		cache.PreRenderredNumsTextures[char] = CharTexture{texture, numsSurface.W}
    }

	var (
		w, h    = e.window.GetSize()
		rows    = h / height
		columns = w / (mn + e.font.GetSpaceBetween())
	)
	cache.RectangleMatrix = NewRectangleMatrix(rows, columns, e.font.GetSize(), e.font.GetSpaceBetween())
	e.cache = cache

	return nil
}

func (e *Engine) LoadTextFromFile(cursorId int64) error {
    str, err := filemanager.ReadFromFile(e.filename)
    if err != nil {
        return err
    }
    println(str)
    err = e.text.Paste(str, cursorId)
    if err != nil {
        return err
    }
    e.text.SetCursorStartPosition(cursorId)
    return nil
}

func (e *Engine) SaveTextToFile() error {
    println(e.text.GetString())
    return filemanager.SaveToFile(e.filename, e.text.GetString())
}

// TODO подумать о том, чтобы не ифать два раза а в EraseChar, InsertChar, MoveCursor передавать просто scancode
func (e *Engine) Loop() {
	running := true
    DEBUG_CUR_ID := int64(0)
    err := e.LoadTextFromFile(DEBUG_CUR_ID)
    if err != nil {
        println(err)
    }
	e.renderText(DEBUG_CUR_ID)
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				//println("Quit")
				running = false
				break
			case *sdl.TextInputEvent:
				pressedKey := t.GetText()
				e.InsertChar(rune(pressedKey[0]), DEBUG_CUR_ID)
				e.renderText(DEBUG_CUR_ID)
				break
			case *sdl.KeyboardEvent:
				// this branch active too when *sdl.TextInputEvent
				// TODO be careful!
				if t.State != sdl.PRESSED {
					break
				}
                //println("Scancode: ", t.Keysym.Scancode)

                key := t.Keysym.Scancode

                if key == sdl.SCANCODE_DELETE || key == sdl.SCANCODE_BACKSPACE {
                    e.EraseChar(key, DEBUG_CUR_ID)
                }

                if key == sdl.SCANCODE_LEFT || key == sdl.SCANCODE_RIGHT || key == sdl.SCANCODE_UP || key == sdl.SCANCODE_DOWN || key == sdl.SCANCODE_HOME || key == 95 || key == sdl.SCANCODE_END || key == 89 {
                    e.MoveCursor(key, DEBUG_CUR_ID)
                }

				switch t.Keysym.Scancode {
				case sdl.SCANCODE_RETURN:
					e.InsertChar('\n', DEBUG_CUR_ID)

				case sdl.SCANCODE_TAB:
					// TODO 4 -> SPACE_IN_ONE_TAB
					e.InsertChar('\t', DEBUG_CUR_ID)
                }
				e.renderText(DEBUG_CUR_ID)
				break
			}
		}
		sdl.Delay(50)
	}

    err = e.SaveTextToFile()
    if err != nil {
        println(err)
    }
}

func (e *Engine) MoveCursor(direction sdl.Scancode, cursorId int64) {
	cur := e.text.Cursors[cursorId]
    switch direction {
    case sdl.SCANCODE_LEFT:
		cur.MoveLeft()
    case sdl.SCANCODE_RIGHT:
		cur.MoveRight()
    case sdl.SCANCODE_UP:
		cur.MoveUp()
    case sdl.SCANCODE_DOWN:
		cur.MoveDown()
    case sdl.SCANCODE_HOME:
        cur.MoveHome()
    case 95:
        cur.MoveHome()
    case sdl.SCANCODE_END:
        cur.MoveEnd()
    case 89:
        cur.MoveEnd()
    }
}

func (e *Engine) EraseChar(key sdl.Scancode, cursorId int64) {
    var err  error
    switch key {
    case sdl.SCANCODE_BACKSPACE:
        err = e.text.RemoveCharBefore(cursorId)
    case sdl.SCANCODE_DELETE:
        err = e.text.RemoveCharAfter(cursorId)
    }
    if err != nil {
        println(err)
    }
}

// TODO ифать \n и \t внутри самого InsertCharAfter!!!
func (e *Engine) InsertChar(value rune, cursorId int64) {
	//_ := e.text.Cursors[cursorId]
	if value == '\n' {
		err := e.text.InsertLineAfter(cursorId)
		if err != nil {
			//println(err)
		}
	} else if value == '\t' {
		for i := 0; i < 4; i++ {
			e.InsertChar(' ', cursorId)
		}

	} else {
		//if cur.LineIter.GetValue().Length()+1 < e.cache.RectangleMatrix.Columns {
		err := e.text.InsertCharAfter(cursorId, value)
		if err != nil {
			//println(err)
		}
		//}
	}
}

func (e *Engine) renderText(cursorId int64) {
	e.renderer.Clear()
	var (
        // e.font.GetSpaceBetween равен 0, поэтому он сейчас ни на что не влияет || тут +7 px это просто отступ от первой цифры
        PaddingLeft int32 = (e.font.GetSpaceBetween() + e.cache.PreRenderredCharTextures[rune('1')].Width) * 4 + 7
		X   int32 = PaddingLeft
		Y   int32 = 0
		row int32 = 0
        // тут col равен 4 потому что мы 4 колоноки оставляем под нумерацию 
		col int32 = 4
        // TODO относительные координаты курсора от экрана

        delta int32 = e.text.Cursors[cursorId].Col - e.cache.RectangleMatrix.Columns + 2 + 4
	)

    //println("delta:",  delta)

    leftBorder := e.text.Cursors[cursorId].ScreenLeft
    //println("LeftBorderText: ", leftBorder)
    if e.text.Cursors[cursorId].Col < leftBorder {
        e.text.Cursors[cursorId].ScreenLeft = e.text.Cursors[cursorId].Col + 1
        delta = e.text.Cursors[cursorId].ScreenLeft
    } else if delta >= leftBorder {
        e.text.Cursors[cursorId].ScreenLeft = delta
    } else {
        delta = leftBorder
    }

    if delta < 0 {
        delta = 0
    }

	for _, c := range e.text.GetScreenString(0) {
        if col - 4 < delta {
            if c == '\n' {
                Y += e.font.GetSize()
                X = PaddingLeft + e.font.GetSpaceBetween()
                row++
                col = 4
            } else {
                col++
            }
            continue
        }
        if col - delta + 1 >= e.cache.RectangleMatrix.Columns {
            if c == '\n' {
                Y += e.font.GetSize()
                X = PaddingLeft + e.font.GetSpaceBetween()
                row++
                col = 4
            } else {
                col++
            }
            continue
        }
        //println("col - del: ", col - delta, "c:", rune(c))
		e.GetRectFromMatrix(row, col - delta).H = e.font.GetSize()
		e.GetRectFromMatrix(row, col - delta).W = e.cache.PreRenderredCharTextures[rune(c)].Width
        e.GetRectFromMatrix(row, col - delta).X = X
		e.GetRectFromMatrix(row, col - delta).Y = Y
		e.renderer.Copy(e.cache.PreRenderredCharTextures[rune(c)].Texture, nil, e.GetRectFromMatrix(row, col - delta))
		X += e.cache.PreRenderredCharTextures[rune(c)].Width + e.font.GetSpaceBetween()
		col++
		if c == '\n' {
			Y += e.font.GetSize()
			X = PaddingLeft + e.font.GetSpaceBetween()
			row++
			col = 4
		}
	}

    cur := e.text.Cursors[cursorId]
    row = 0
    col = 3
    X = (e.font.GetSpaceBetween() + e.cache.PreRenderredNumsTextures[rune('1')].Width) * 3
    Y = 0
    for num := cur.ScreenHead.RowNumber; num <= cur.ScreenTail.RowNumber; num++ {
        for cp := num; cp > 0; cp /= 10 {
            if col < 0 {
                println("Слишком длиный номер строки, лучше переписать код на микросервисы")
                break
            }
            e.GetRectFromMatrix(row, col).H = e.font.GetSize()
            e.GetRectFromMatrix(row, col).W = e.cache.PreRenderredNumsTextures[DtoR(cp % 10)].Width
            e.GetRectFromMatrix(row, col).X = X
            e.GetRectFromMatrix(row, col).Y = Y
            e.renderer.Copy(e.cache.PreRenderredNumsTextures[DtoR(cp % 10)].Texture, nil, e.GetRectFromMatrix(row, col))
            X -= e.cache.PreRenderredNumsTextures[DtoR(cp % 10)].Width + e.font.GetSpaceBetween()

            col--

        }
        if num == 0 {
            e.GetRectFromMatrix(row, col).H = e.font.GetSize()
            e.GetRectFromMatrix(row, col).W = e.cache.PreRenderredNumsTextures[rune(48)].Width
            e.GetRectFromMatrix(row, col).X = X
            e.GetRectFromMatrix(row, col).Y = Y
            e.renderer.Copy(e.cache.PreRenderredNumsTextures[DtoR(0)].Texture, nil, e.GetRectFromMatrix(row, col))

        }
        Y += e.font.GetSize()
        X = (e.font.GetSpaceBetween() + e.cache.PreRenderredNumsTextures[rune('1')].Width) * 3
        row++
        col = 3
    }

	e.RenderCursor(cursorId)
	e.renderer.Present()

}

// digit to rune
func DtoR(n int32) rune {
    if 0 <= n && n <= 9 {
        return rune(48 + n)
    }
    return rune('0')
}

func main() {
	var (
		ScreenHeight    int32  = 980
		ScreenWidth     int32  = 1480
		FontSize        int32  = 52
		SpaceBetween    int32  = 0
		FontFilename    string = "MonoNL-Regular.ttf"
		FontColor       uint32 = 0xFFFFFFFF
		NumbersColor    uint32 = 0x34C136FF
		WindowTitle     string = "Type2gether"
	)

	GUIStart()
	defer GUIStop()

	engine, err := NewEngine(ScreenWidth, ScreenHeight, FontFilename, FontSize, SpaceBetween, hexToSdlColor(FontColor), hexToSdlColor(NumbersColor), WindowTitle, AllSupportedChars)

	if err != nil {
		panic(err)
	}

	engine.Loop()
}

