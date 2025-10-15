package csvi

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"

	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/keys"

	"github.com/hymkor/csvi/candidate"
	"github.com/hymkor/csvi/uncsv"

	"github.com/hymkor/csvi/internal/ansi"
	"github.com/hymkor/csvi/internal/manualctl"
	"github.com/hymkor/csvi/internal/nonblock"
)

type colorSet struct {
	On  string
	Off string
	Rev string
}

func (c *colorSet) Revert() {
	if c.Rev != "" {
		tmp := c.On
		c.On = c.Rev
		c.Rev = tmp
	}
}

type _ColorStyle struct {
	Cursor, Even, Odd colorSet
}

func (v *_ColorStyle) Revert() {
	v.Cursor.Revert()
	v.Even.Revert()
	v.Odd.Revert()
}

var bodyColorStyle = _ColorStyle{
	Cursor: colorSet{On: "\x1B[107;30;22m", Off: "\x1B[49;39m", Rev: "\x1B[40;37m"},
	Even:   colorSet{On: "\x1B[48;5;235;39;1m", Off: "\x1B[22;49m", Rev: "\x1B[48;5;252;39m"},
	Odd:    colorSet{On: "\x1B[49;39;1m", Off: "\x1B[22m"},
}

var headColorStyle = _ColorStyle{
	Cursor: colorSet{On: "\x1B[107;30;22m", Off: "\x1B[49;36m", Rev: "\x1B[40;36m"},
	Even:   colorSet{On: "\x1B[48;5;235;36;1m", Off: "\x1B[22;49m", Rev: "\x1B[48;5;252;36m"},
	Odd:    colorSet{On: "\x1B[49;36;1m", Off: "\x1B[22m"},
}

var monoChromeStyle = _ColorStyle{
	Cursor: colorSet{On: "\x1B[7m", Off: "\x1B[0m"},
	Even:   colorSet{On: "\x1B[0m", Off: "\x1B[0m"},
	Odd:    colorSet{On: "\x1B[0m", Off: "\x1B[0m"},
}

func RevertColor() {
	bodyColorStyle.Revert()
	headColorStyle.Revert()
}

func MonoChrome() {
	bodyColorStyle = monoChromeStyle
	headColorStyle = monoChromeStyle
	ansi.YELLOW = ""
}

var replaceTable = strings.NewReplacer(
	"\r", "\u240D",
	"\x1B", "\u241B",
	"\n", "\u240A",
	"\t", "\u2409")

// See. en.wikipedia.org/wiki/Unicode_control_characters#Control_pictures

func sum(f func(n int) int, from, to int) int {
	result := 0
	for i := from; i < to; i++ {
		result += f(i)
	}
	return result
}

func drawLine(
	csvs []uncsv.Cell,
	cellWidth func(int) int,
	screenWidth int,
	cursorPos int,
	reverse bool,
	style *_ColorStyle,
	out io.Writer) {

	if len(csvs) <= 0 && cursorPos >= 0 {
		io.WriteString(out, style.Cursor.On)
		io.WriteString(out, "\x1B[K")
		io.WriteString(out, style.Cursor.Off)
		return
	}
	i := 0

	if reverse {
		io.WriteString(out, style.Odd.On)
		defer io.WriteString(out, style.Odd.Off)
	} else {
		io.WriteString(out, style.Even.On)
		defer io.WriteString(out, style.Even.Off)
	}
	io.WriteString(out, "\x1B[K")

	for len(csvs) > 0 {
		cursor := csvs[0]
		text := cursor.Text()
		csvs = csvs[1:]
		nextI := i + 1

		cw := cellWidth(i)
		for len(csvs) > 0 && csvs[0].Text() == "" && nextI != cursorPos {
			cw += cellWidth(nextI)
			csvs = csvs[1:]
			nextI++
		}
		if cw > screenWidth || len(csvs) <= 0 {
			cw = screenWidth
		}
		text = replaceTable.Replace(text)
		ss, _ := cutStrInWidth(text, cw)
		if i == cursorPos {
			io.WriteString(out, style.Cursor.On)
		}
		if cursor.Modified() {
			io.WriteString(out, ansi.UNDERLINE_ON)
		}
		io.WriteString(out, ss)
		if cursor.Modified() {
			io.WriteString(out, ansi.UNDERLINE_OFF)
		}
		if i == cursorPos {
			io.WriteString(out, "\x1B[K")
			if reverse {
				io.WriteString(out, style.Odd.On)
			} else {
				io.WriteString(out, style.Even.On)
			}
		}
		screenWidth -= cw
		if screenWidth <= 0 {
			break
		}
		fmt.Fprintf(out, "\x1B[%dG", sum(cellWidth, 0, nextI)+1)
		if i == cursorPos {
			io.WriteString(out, "\x1B[K")
		}
		i = nextI
	}
}

func up(n int, out io.Writer) {
	if n == 0 {
		out.Write([]byte{'\r'})
	} else if n == 1 {
		io.WriteString(out, "\r\x1B[A")
	} else {
		fmt.Fprintf(out, "\r\x1B[%dA", n)
	}
}

func drawPage(page func(func([]uncsv.Cell) bool), cellWidth func(int) int, csrpos, csrlin, w, h int, style *_ColorStyle, cache map[int]string, out io.Writer) int {
	reverse := false
	count := 0
	lfCount := 0
	page(func(record []uncsv.Cell) bool {
		if count >= h {
			return false
		}
		if count > 0 {
			lfCount++
			io.WriteString(out, "\r\n") // "\r" is for Linux and go-tty
		}
		cursorPos := -1
		if count == csrlin {
			cursorPos = csrpos
		}
		var buffer strings.Builder
		drawLine(record, cellWidth, w, cursorPos, reverse, style, &buffer)
		line := buffer.String()
		if f := cache[count]; f != line {
			io.WriteString(out, line)
			cache[count] = line
		}
		reverse = !reverse
		count++
		return true
	})
	io.WriteString(out, "\r\n") // \r is for Linux & go-tty
	lfCount++
	return lfCount
}

func cellsAfter(cells []uncsv.Cell, n int) []uncsv.Cell {
	if n <= len(cells) {
		return cells[n:]
	} else {
		return []uncsv.Cell{}
	}
}

type _View struct {
	headCache map[int]string
	bodyCache map[int]string
}

func newView() *_View {
	return &_View{
		headCache: map[int]string{},
		bodyCache: map[int]string{},
	}
}

func (v *_View) clearCache() {
	for k := range v.headCache {
		delete(v.headCache, k)
	}
	for k := range v.bodyCache {
		delete(v.bodyCache, k)
	}
}

func (v *_View) Draw(header, startRow, cursorRow *RowPtr, _cellWidth *CellWidth, headerLines, startCol, cursorCol, screenHeight, screenWidth int, out io.Writer) int {
	// print header
	lfCount := 0
	cellWidth := func(n int) int {
		return _cellWidth.Get(n + startCol)
	}
	if h := headerLines; h > 0 {
		enum := func(callback func([]uncsv.Cell) bool) {
			for i := 0; i < h && header != nil; i++ {
				if !callback(cellsAfter(header.Cell, startCol)) {
					return
				}
				header = header.Next()
			}
		}
		lfCount = drawPage(enum, cellWidth, cursorCol-startCol, cursorRow.lnum, screenWidth-1, h, &headColorStyle, v.headCache, out)
	}
	if startRow.lnum < headerLines {
		for i := 0; i < headerLines && startRow != nil; i++ {
			startRow = startRow.Next()
		}
	}
	if startRow == nil {
		return lfCount
	}
	p := startRow.Clone()
	// print body
	enum := func(callback func([]uncsv.Cell) bool) {
		for p != nil {
			if !callback(cellsAfter(p.Cell, startCol)) {
				return
			}
			p = p.Next()
		}
	}
	style := &bodyColorStyle
	if headerLines%2 == 1 {
		style = &_ColorStyle{
			Cursor: bodyColorStyle.Cursor,
			Even:   bodyColorStyle.Odd,
			Odd:    bodyColorStyle.Even,
		}
	}
	return lfCount + drawPage(enum, cellWidth, cursorCol-startCol, cursorRow.lnum-startRow.lnum, screenWidth-1, screenHeight-1, style, v.bodyCache, out)
}

func (app *_Application) MessageAndGetKey(message string) (string, error) {
	fmt.Fprintf(app, "%s\r%s%s", ansi.YELLOW, message, ansi.ERASE_LINE)
	io.WriteString(app, ansi.CURSOR_ON)
	ch, err := app.GetKey()
	io.WriteString(app, ansi.CURSOR_OFF)
	return ch, err
}

func (app *_Application) YesNo(message string) bool {
	ch, err := app.MessageAndGetKey(message)
	return err == nil && ch == "y"
}

func first[T any](value T, _ error) T {
	return value
}

func printStatusLine(out io.Writer, mode *uncsv.Mode, cursorRow *RowPtr, cursorCol int, screenWidth int) {
	n := 0
	if mode.Comma == '\t' {
		n += first(io.WriteString(out, "[TSV]"))
	} else if mode.Comma == ',' {
		n += first(io.WriteString(out, "[CSV]"))
	}
	switch cursorRow.Term {
	case "\r\n":
		n += first(io.WriteString(out, "[CRLF]"))
	case "\n":
		n += first(io.WriteString(out, "[LF]"))
	case "":
		n += first(io.WriteString(out, "[EOF]"))
	}
	if mode.HasBom() {
		n += first(io.WriteString(out, "[BOM]"))
	}
	if mode.NonUTF8 {
		if mode.IsUTF16LE() {
			n += first(io.WriteString(out, "[16LE]"))
		} else if mode.IsUTF16BE() {
			n += first(io.WriteString(out, "[16BE]"))
		} else {
			n += first(io.WriteString(out, "[ANSI]"))
		}
	}
	if 0 <= cursorCol && cursorCol < len(cursorRow.Cell) {
		n += first(fmt.Fprintf(out, "(%d,%d/%d): ",
			cursorCol+1,
			cursorRow.lnum+1,
			cursorRow.list.Len()))
		var buffer strings.Builder
		buffer.WriteString(cursorRow.Cell[cursorCol].SourceText(mode))
		if cursorCol < len(cursorRow.Cell)-1 {
			buffer.WriteByte(mode.Comma)
		} else if term := cursorRow.Term; term != "" {
			buffer.WriteString(term)
		} else { // EOF
			buffer.WriteString("\u2592")
		}
		io.WriteString(out, runewidth.Truncate(replaceTable.Replace(buffer.String()), screenWidth-n, "..."))
	}
}

type Pilot interface {
	Size() (int, int, error)
	GetKey() (string, error)
	ReadLine(out io.Writer, prompt, defaultText string, c candidate.Candidate) (string, error)
	GetFilename(io.Writer, string, string) (string, error)
	Close() error
}

type CommandResult struct {
	Message string
	Quit    bool
}

type CellValidatedEvent struct {
	Text string
	Row  int
	Col  int
}

type KeyEventArgs struct {
	*_Application
	CursorRow *RowPtr
	CursorCol int
}

type Config struct {
	*uncsv.Mode
	CellWidth   *CellWidth
	HeaderLines int
	Pilot
	FixColumn       bool
	ReadOnly        bool
	ProtectHeader   bool
	Message         string
	KeyMap          map[string]func(*KeyEventArgs) (*CommandResult, error)
	OnCellValidated func(*CellValidatedEvent) (string, error)
	Titles          []string
}

func (cfg Config) validate(row *RowPtr, col int, text string) (string, error) {
	if cfg.OnCellValidated == nil {
		return text, nil
	}
	return cfg.OnCellValidated(&CellValidatedEvent{
		Row:  row.lnum,
		Col:  col,
		Text: text,
	})
}

func (cfg Config) Edit(dataSource io.Reader, ttyOut io.Writer) (*Result, error) {
	if dataSource == nil {
		return cfg.edit(nil, ttyOut)
	}
	bufDataSource, ok := dataSource.(*bufio.Reader)
	if !ok {
		bufDataSource = bufio.NewReader(dataSource)
	}
	return cfg.edit(func() (*uncsv.Row, error) {
		return uncsv.ReadLine(bufDataSource, cfg.Mode)
	}, ttyOut)
}

func (cfg Config) EditFromStringSlice(fetch func() ([]string, bool), ttyOut io.Writer) (*Result, error) {
	return cfg.edit(func() (*uncsv.Row, error) {
		slice, ok := fetch()
		if !ok {
			return nil, io.EOF
		}
		row := uncsv.NewRowFromStringSlice(cfg.Mode, slice)
		return &row, nil
	}, ttyOut)
}

func isEmptyRow(row *uncsv.Row) bool {
	switch len(row.Cell) {
	case 0:
		return true
	case 1:
		if len(row.Cell[0].Original()) <= 0 {
			return true
		}
	}
	return false
}

const (
	msgReadOnly      = "Read Only Mode !"
	msgProtectHeader = "Header is protected"
	msgColumnFixed   = "The order of Columns is fixed !"
)

func (cfg *Config) checkWriteProtect(cursorRow *RowPtr) string {
	if cfg.ProtectHeader && cursorRow.lnum < cfg.HeaderLines {
		return msgProtectHeader
	}
	if cfg.ReadOnly {
		return msgReadOnly
	}
	return ""
}

func (cfg *Config) checkWriteProtectAndColumn(cursorRow *RowPtr) string {
	if m := cfg.checkWriteProtect(cursorRow); m != "" {
		return m
	}
	if cfg.FixColumn {
		return msgColumnFixed
	}
	return ""
}

func (app *_Application) readlineAndValidate(prompt, text string, row *RowPtr, col int) (string, error) {
	candidates := makeCandidate(row.lnum-1, col, row)
	for {
		var err error
		text, err = app.Config.Pilot.ReadLine(app.out, prompt, text, candidates)
		if err != nil {
			return "", err
		}
		tx, err := app.Config.validate(row, col, text)
		if err == nil {
			return tx, nil
		}
		prompt = fmt.Sprintf("%s: Re-enter>", err.Error())
	}
}

func (cfg *Config) edit(fetch func() (*uncsv.Row, error), out io.Writer) (*Result, error) {
	if cfg.KeyMap == nil {
		cfg.KeyMap = make(map[string]func(*KeyEventArgs) (*CommandResult, error))
	}

	mode := cfg.Mode
	if mode == nil {
		mode = &uncsv.Mode{}
	}

	cellWidth := cfg.CellWidth
	if cellWidth == nil {
		cellWidth = NewCellWidth()
	}

	pilot := cfg.Pilot
	if pilot == nil {
		var err error
		pilot, err = manualctl.New()
		if err != nil {
			return nil, err
		}
		defer pilot.Close()
		cfg.Pilot = pilot
	}
	app := &_Application{
		Config:   cfg,
		csvLines: list.New(),
		out:      out,
	}
	if fetch != nil {
		for i := 0; i < 100; i++ {
			row, err := fetch()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				fetch = nil
				if isEmptyRow(row) {
					break
				}
			}
			app.Push(row)
			if err == io.EOF {
				break
			}
		}
	}
	startRow := app.Front()
	startCol := 0
	if startRow == nil {
		newRow := uncsv.NewRow(mode)
		app.Push(&newRow)
		startRow = app.Front()
	}
	cursorCol := 0
	cursorRow := app.Front()

	lastSearch := searchForward
	lastSearchRev := searchBackward
	lastWord := ""
	var lastWidth, lastHeight int

	keyWorker := nonblock.New(func() (string, error) {
		return pilot.GetKey()
	})
	defer keyWorker.Close()

	view := newView()

	screenWidth, _screenHeight, err := pilot.Size()
	if err != nil {
		return nil, err
	}
	for _, title := range cfg.Titles {
		s, _ := cutStrInWidth(title, screenWidth-1)
		fmt.Fprintln(out, s)
	}
	message := cfg.Message
	var killbuffer string
	for {
		screenHeight := _screenHeight - len(cfg.Titles)
		screenHeight -= cfg.HeaderLines
		if lastWidth != screenWidth || lastHeight != screenHeight {
			view.clearCache()
			lastWidth = screenWidth
			lastHeight = screenHeight
			io.WriteString(out, ansi.CURSOR_OFF)
		}

		lfCount := view.Draw(app.Front(), startRow, cursorRow, cellWidth, cfg.HeaderLines, startCol, cursorCol, screenHeight, screenWidth, out)
		repaint := func() {
			up(lfCount, out)
			lfCount = view.Draw(app.Front(), startRow, cursorRow, cellWidth, cfg.HeaderLines, startCol, cursorCol, screenHeight, screenWidth, out)
		}

		io.WriteString(out, ansi.YELLOW)
		if message != "" {
			io.WriteString(out, runewidth.Truncate(message, screenWidth-1, ""))
		} else if 0 <= cursorRow.lnum && cursorRow.lnum < app.Len() {
			printStatusLine(out, mode, cursorRow, cursorCol, screenWidth)
		}
		io.WriteString(out, ansi.RESET)
		io.WriteString(out, ansi.ERASE_SCRN_AFTER)

		const interval = 4
		displayUpdateTime := time.Now().Add(time.Second / interval)

		ch, err := keyWorker.GetOr(func() bool {
			if fetch == nil {
				return false
			}
			row, err := fetch()
			if err != nil {
				fetch = nil
				if err != io.EOF || isEmptyRow(row) {
					return false
				}
			}
			app.Push(row)
			if message == "" && (err == io.EOF || time.Now().After(displayUpdateTime)) {
				io.WriteString(out, "\r"+ansi.YELLOW)
				printStatusLine(out, mode, cursorRow, cursorCol, screenWidth)
				io.WriteString(out, ansi.RESET)
				io.WriteString(out, ansi.ERASE_SCRN_AFTER)
				displayUpdateTime = time.Now().Add(time.Second / interval)
			}
			return err != io.EOF
		})
		if err != nil {
			return nil, err
		}
		message = ""

		if handler, ok := cfg.KeyMap[ch]; ok {
			e := &KeyEventArgs{
				CursorRow:    cursorRow,
				CursorCol:    cursorCol,
				_Application: app,
			}
			cmdResult, err := handler(e)
			if err != nil || cmdResult.Quit {
				return &Result{_Application: app}, err
			}
			message = cmdResult.Message
		} else {
			switch ch {
			case keys.CtrlL:
				view.clearCache()
			case "L":
				newEncoding, err := pilot.ReadLine(out, "This will discard all unsaved changes. Switch encoding to:", "", ianaNames)
				if err != nil {
					message = err.Error()
					break
				}
				if err := mode.SetEncoding(newEncoding); err != nil {
					message = err.Error()
					break
				}
				mode.NonUTF8 = true
				cfg.Mode = mode
				for p := app.Front(); p != nil; p = p.Next() {
					p.Restore(mode)
				}
				view.clearCache()
			case "q", keys.Escape:
				if cfg.ReadOnly || app.YesNo("Quit Sure ? [y/n]") {
					io.WriteString(out, "\n")
					return &Result{_Application: app}, nil
				}
			case "j", keys.Down, keys.CtrlN, keys.Enter:
				if next := cursorRow.Next(); next != nil {
					cursorRow = next
				}
			case "k", keys.Up, keys.CtrlP:
				if prev := cursorRow.Prev(); prev != nil {
					cursorRow = prev
				}
			case "h", keys.Left, keys.CtrlB, keys.ShiftTab:
				if cursorCol > 0 {
					cursorCol--
				}
			case "l", keys.Right, keys.CtrlF, keys.CtrlI:
				cursorCol++
			case "0", "^", keys.CtrlA:
				cursorCol = 0
			case "$", keys.CtrlE:
				cursorCol = len(cursorRow.Cell) - 1
			case "<":
				cursorRow = app.Front()
				startRow = app.Front()
				cursorCol = 0
				startCol = 0
			case ">", "G":
				cursorRow = app.Back()
			case "n":
				if lastWord == "" {
					break
				}
				r, c := lastSearch(cursorRow, cursorCol, lastWord)
				if r == nil {
					message = fmt.Sprintf("%s: not found", lastWord)
					break
				}
				cursorRow = r
				cursorCol = c
			case "N":
				if lastWord == "" {
					break
				}
				r, c := lastSearchRev(cursorRow, cursorCol, lastWord)
				if r == nil {
					message = fmt.Sprintf("%s: not found", lastWord)
					break
				}
				cursorRow = r
				cursorCol = c
			case "*", "#":
				view.clearCache()
				if ch == "*" || ch == "#" {
					lastWord = cursorRow.Cell[cursorCol].Text()
				}
				if ch == "*" {
					lastSearch = searchExactForward
					lastSearchRev = searchExactBackward
				} else {
					lastSearch = searchExactBackward
					lastSearchRev = searchExactForward
				}
				r, c := lastSearch(cursorRow, cursorCol, lastWord)
				if r == nil {
					message = fmt.Sprintf("%s: not found", lastWord)
					break
				}
				cursorRow = r
				cursorCol = c
			case "/", "?":
				var err error
				view.clearCache()
				lastWord, err = pilot.ReadLine(out, ch, "", nil)
				if err != nil {
					if err != readline.CtrlC {
						message = err.Error()
					}
					break
				}
				if ch == "/" {
					lastSearch = searchForward
					lastSearchRev = searchBackward
				} else {
					lastSearch = searchBackward
					lastSearchRev = searchForward
				}
				r, c := lastSearch(cursorRow, cursorCol, lastWord)
				if r == nil {
					message = fmt.Sprintf("%s: not found", lastWord)
					break
				}
				cursorRow = r
				cursorCol = c
			case "o":
				if cfg.ProtectHeader && cursorRow.lnum+1 < cfg.HeaderLines {
					message = msgProtectHeader
					break
				}
				if cfg.ReadOnly {
					message = msgReadOnly
					break
				}
				newRow := uncsv.NewRow(mode)
				newRow.Term = cursorRow.Term
				if cursorRow.Term == "" {
					cursorRow.Term = mode.DefaultTerm
				}
				if cfg.FixColumn {
					for len(newRow.Cell) < len(cursorRow.Cell) {
						newRow.Insert(0, "", mode)
					}
				}
				cursorRow = cursorRow.InsertAfter(&newRow)
				repaint()
				view.clearCache()
				newCol := cursorCol
				if cursorCol >= len(cursorRow.Cell) {
					newCol = len(cursorRow.Cell) - 1
				}
				if text, err := app.readlineAndValidate("new line>", "", cursorRow, newCol); err == nil {
					cursorRow.Replace(newCol, text, mode)
				}
			case "O":
				if m := cfg.checkWriteProtect(cursorRow); m != "" {
					message = m
					break
				}
				startPrevP := startRow.Prev()
				newRow := uncsv.NewRow(mode)
				if cfg.FixColumn {
					for len(newRow.Cell) < len(cursorRow.Cell) {
						newRow.Insert(0, "", mode)
					}
				}
				cursorRow = cursorRow.InsertBefore(&newRow)
				if startPrevP != nil {
					startRow = startPrevP.Next()
				} else {
					startRow = app.Front()
				}
				repaint()
				view.clearCache()
				newCol := cursorCol
				if cursorCol >= len(cursorRow.Cell) {
					newCol = len(cursorRow.Cell) - 1
				}
				if text, err := app.readlineAndValidate("new line>", "", cursorRow, newCol); err == nil {
					cursorRow.Replace(newCol, text, mode)
				}
			case "D":
				if m := cfg.checkWriteProtect(cursorRow); m != "" {
					message = m
					break
				}
				if app.Len() <= 1 {
					break
				}
				startPrevP := startRow.Prev()
				prevP := cursorRow.Prev()
				removedRow := cursorRow.Remove()
				app.removedRows = append(app.removedRows, removedRow)
				if prevP == nil {
					cursorRow = app.Front()
				} else if next := prevP.Next(); next != nil {
					cursorRow = next
				} else {
					cursorRow = prevP
					cursorRow.Term = removedRow.Term
				}
				if startPrevP == nil {
					startRow = app.Front()
				} else {
					startRow = startPrevP.Next()
				}
			case "i":
				if m := cfg.checkWriteProtectAndColumn(cursorRow); m != "" {
					message = m
					break
				}
				view.clearCache()
				if text, err := app.readlineAndValidate("insert cell>", "", cursorRow, cursorCol); err == nil {
					if cells := cursorRow.Cell; len(cells) == 1 && cells[0].Text() == "" {
						cursorRow.Replace(cursorCol, text, mode)
					} else {
						cursorRow.Insert(cursorCol, text, mode)
						cursorCol++
					}
				}
			case "a":
				if m := cfg.checkWriteProtectAndColumn(cursorRow); m != "" {
					message = m
					break
				}
				if cells := cursorRow.Cell; len(cells) == 1 && cells[0].Text() == "" {
					// current column is the last one and it is empty
					view.clearCache()
					if text, err := app.readlineAndValidate("append cell>", "", cursorRow, cursorCol); err == nil {
						cursorRow.Replace(cursorCol, text, mode)
					}
				} else {
					cursorCol++
					cursorRow.Insert(cursorCol, "", mode)
					repaint()
					view.clearCache()
					if text, err := app.readlineAndValidate("append cell>", "", cursorRow, cursorCol); err != nil {
						// cancel
						cursorRow.Delete(cursorCol)
						cursorCol--
					} else {
						cursorRow.Replace(cursorCol, text, mode)
					}
				}
			case "r", "R", keys.F2:
				if m := cfg.checkWriteProtect(cursorRow); m != "" {
					message = m
					break
				}
				cursor := &cursorRow.Cell[cursorCol]
				q := cursor.IsQuoted()
				view.clearCache()
				if text, err := app.readlineAndValidate("replace cell>", cursor.Text(), cursorRow, cursorCol); err == nil {
					cursorRow.Replace(cursorCol, text, mode)
					if q {
						*cursor = cursor.Quote(mode)
					}
				}
			case "u":
				cursorRow.Cell[cursorCol].Restore(mode)
			case "y":
				killbuffer = cursorRow.Cell[cursorCol].Text()
				message = "yanked the current cell: " + killbuffer
			case "p":
				if m := cfg.checkWriteProtect(cursorRow); m != "" {
					message = m
					break
				}
				cursorRow.Replace(cursorCol, killbuffer, mode)
				message = "pasted: " + killbuffer
			case "d", "x":
				if m := cfg.checkWriteProtectAndColumn(cursorRow); m != "" {
					message = m
					break
				}
				if len(cursorRow.Cell) <= 1 {
					cursorRow.Replace(0, "", mode)
				} else {
					cursorRow.Delete(cursorCol)
				}
			case "\"":
				cursor := &cursorRow.Cell[cursorCol]
				if cursor.IsQuoted() {
					cursorRow.Replace(cursorCol, cursor.Text(), mode)
				} else {
					*cursor = cursor.Quote(mode)
				}
			case "w":
				if fetch != nil {
					io.WriteString(out, ansi.YELLOW+"\rw: Wait a moment for reading all data..."+ansi.ERASE_LINE)
					for {
						row, err := fetch()
						if err != nil && err != io.EOF {
							return nil, err
						}
						app.Push(row)
						if err == io.EOF {
							break
						}
					}
				}
				if err := cmdWrite(app); err != nil {
					message = err.Error()
				}
				view.clearCache()
			case "]":
				if w := cellWidth.Get(cursorCol); w < 40 {
					cellWidth.Set(cursorCol, w+1)
				}
				view.clearCache()
			case "[":
				if w := cellWidth.Get(cursorCol) - 1; w > 3 {
					cellWidth.Set(cursorCol, w)
				}
				view.clearCache()
			}
		}
		if L := len(cursorRow.Cell); L <= 0 {
			cursorCol = 0
		} else if cursorCol >= L {
			cursorCol = L - 1
		}
		if cursorRow.lnum < startRow.lnum {
			startRow = cursorRow.Clone()
		} else if cursorRow.lnum >= startRow.lnum+screenHeight-1 {
			goal := cursorRow.lnum - (screenHeight - 1) + 1
			for startRow = cursorRow.Clone(); startRow.lnum > goal; {
				startRow = startRow.Prev()
			}
		}
		if cursorCol < startCol {
			startCol = cursorCol
		} else {
			cellWidth := cfg.CellWidth
			if cellWidth == nil {
				cellWidth = NewCellWidth()
			}
			for {
				w := sum(cellWidth.Get, startCol, cursorCol+1)
				if w <= screenWidth {
					break
				}
				startCol++
			}
		}
		up(lfCount, out)
	}
}

func IsRevertVideoWithEnv() bool {
	colorFgBg, ok := os.LookupEnv("COLORFGBG")
	if !ok {
		return false
	}
	fgStr, bgStr, ok := strings.Cut(colorFgBg, ";")
	if !ok {
		return false
	}
	fgInt, err := strconv.ParseInt(fgStr, 10, 64)
	if err != nil {
		return false
	}
	bgInt, err := strconv.ParseInt(bgStr, 10, 64)
	if err != nil {
		return false
	}
	return fgInt < bgInt
}
