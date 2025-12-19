package csvi

import (
	"bufio"
	"container/list"
	"errors"
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

type colorStyle struct {
	Cursor, Even, Odd colorSet
}

func (v *colorStyle) Revert() {
	v.Cursor.Revert()
	v.Even.Revert()
	v.Odd.Revert()
}

var bodyColorStyle = colorStyle{
	Cursor: colorSet{On: "\x1B[107;30;22m", Off: "\x1B[49;39m", Rev: "\x1B[40;37m"},
	Even:   colorSet{On: "\x1B[48;5;235;39;1m", Off: "\x1B[22;49m", Rev: "\x1B[48;5;252;39m"},
	Odd:    colorSet{On: "\x1B[49;39;1m", Off: "\x1B[22m"},
}

var headColorStyle = colorStyle{
	Cursor: colorSet{On: "\x1B[107;30;22m", Off: "\x1B[49;36m", Rev: "\x1B[40;36m"},
	Even:   colorSet{On: "\x1B[48;5;235;36;1m", Off: "\x1B[22;49m", Rev: "\x1B[48;5;252;36m"},
	Odd:    colorSet{On: "\x1B[49;36;1m", Off: "\x1B[22m"},
}

var monoChromeStyle = colorStyle{
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

type lineStyle struct {
	cellWidth    func(int) int
	screenWidth  int
	screenHeight int
	*colorStyle
	sep string
}

func (style lineStyle) drawLine(
	field []uncsv.Cell,
	cursorPos int,
	reverse bool,
	out io.Writer) {

	if len(field) <= 0 && cursorPos >= 0 {
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

	screenWidth := style.screenWidth
	sepLen := runewidth.StringWidth(style.sep)

	for len(field) > 0 {
		cursor := field[0]
		text := cursor.Text()
		field = field[1:]
		nextI := i + 1

		cw := style.cellWidth(i)
		for len(field) > 0 && field[0].Text() == "" && nextI != cursorPos {
			cw += style.cellWidth(nextI)
			field = field[1:]
			nextI++
		}
		if cw > screenWidth || len(field) <= 0 {
			cw = screenWidth
		}
		text = replaceTable.Replace(text)
		if i > 0 && style.sep != "" {
			io.WriteString(out, "\x1B[30;1m")
			io.WriteString(out, style.sep)
			if reverse {
				io.WriteString(out, style.Odd.On)
			} else {
				io.WriteString(out, style.Even.On)
			}
		}
		text = runewidth.Truncate(text, cw-sepLen, "\u2026")
		if i == cursorPos {
			io.WriteString(out, style.Cursor.On)
		}
		if cursor.Modified() {
			io.WriteString(out, ansi.UNDERLINE_ON)
		}
		io.WriteString(out, text)
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
		fmt.Fprintf(out, "\x1B[%dG", sum(style.cellWidth, 0, nextI)+1)
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

func (style lineStyle) drawPage(page func(func([]uncsv.Cell) bool), csrpos, csrlin int, cache map[int]string, out io.Writer) int {
	reverse := false
	count := 0
	lfCount := 0
	page(func(record []uncsv.Cell) bool {
		if count >= style.screenHeight {
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
		style.drawLine(record, cursorPos, reverse, &buffer)
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

type application struct {
	csvLines     *list.List
	removedRows  []*uncsv.Row
	out          io.Writer
	dirty        int
	lastSavePath string
	startRow     *RowPtr
	cursorRow    *RowPtr
	startCol     int
	cursorCol    int
	screenWidth  int
	screenHeight int
	headCache    map[int]string
	bodyCache    map[int]string
	lfCount      int
	fetch        func() (bool, *uncsv.Row, error)
	tryFetch     func() (bool, *uncsv.Row, error)
	*Config
}

func (cfg *Config) newApplication(out io.Writer) *application {
	return &application{
		headCache: map[int]string{},
		bodyCache: map[int]string{},
		Config:    cfg,
		csvLines:  list.New(),
		out:       out,
	}
}

func (app *application) clearCache() {
	for k := range app.headCache {
		delete(app.headCache, k)
	}
	for k := range app.bodyCache {
		delete(app.bodyCache, k)
	}
}

func (app *application) Rewind() {
	up(app.lfCount, app.out)
}

func (app *application) Repaint() {
	app.Rewind()
	app.Draw()
}

func (app *application) Draw() int {
	// print header
	app.lfCount = 0
	header := app.Front()

	cellWidth := func(n int) int {
		return app.CellWidth.Get(n + app.startCol)
	}
	if h := app.HeaderLines; h > 0 {
		enum := func(callback func([]uncsv.Cell) bool) {
			for i := 0; i < h && header != nil; i++ {
				if !callback(cellsAfter(header.Cell, app.startCol)) {
					return
				}
				header = app.nextOrFetch(header)
			}
		}
		app.lfCount = lineStyle{
			cellWidth:    cellWidth,
			screenWidth:  app.screenWidth - 1,
			screenHeight: h,
			colorStyle:   &headColorStyle,
			sep:          app.OutputSep,
		}.drawPage(enum, app.cursorCol-app.startCol, app.cursorRow.lnum, app.headCache, app.out)
	}
	startRow := app.startRow
	if startRow.lnum < app.HeaderLines {
		for i := 0; i < app.HeaderLines && startRow != nil; i++ {
			startRow = app.nextOrFetch(startRow)
		}
	}
	if startRow == nil {
		return app.lfCount
	}
	p := startRow.Clone()
	// print body
	enum := func(callback func([]uncsv.Cell) bool) {
		for p != nil {
			if !callback(cellsAfter(p.Cell, app.startCol)) {
				return
			}
			p = app.nextOrFetch(p)
		}
	}
	style := &bodyColorStyle
	if app.HeaderLines%2 == 1 {
		style = &colorStyle{
			Cursor: bodyColorStyle.Cursor,
			Even:   bodyColorStyle.Odd,
			Odd:    bodyColorStyle.Even,
		}
	}
	app.lfCount += lineStyle{
		cellWidth:    cellWidth,
		screenWidth:  app.screenWidth - 1,
		screenHeight: app.screenHeight - 1,
		colorStyle:   style,
		sep:          app.OutputSep,
	}.drawPage(enum, app.cursorCol-app.startCol, app.cursorRow.lnum-startRow.lnum, app.bodyCache, app.out)
	return app.lfCount
}

func (app *application) MessageAndGetKey(message string) (string, error) {
	fmt.Fprintf(app, "%s\r%s%s ", ansi.YELLOW, message, ansi.ERASE_LINE)
	io.WriteString(app, ansi.CURSOR_ON)
	ch, err := app.GetKey()
	io.WriteString(app, ansi.CURSOR_OFF)
	return ch, err
}

func (app *application) YesNo(message string) bool {
	ch, err := app.MessageAndGetKey(message)
	return err == nil && ch == "y"
}

func first[T any](value T, _ error) T {
	return value
}

func (app *application) printStatusLine() {
	n := 0
	if app.isDirty() {
		n += first(app.out.Write([]byte{'*'}))
	} else {
		n += first(app.out.Write([]byte{' '}))
	}
	if app.Mode.Comma == '\t' {
		n += first(io.WriteString(app.out, "[TSV]"))
	} else if app.Mode.Comma == ',' {
		n += first(io.WriteString(app.out, "[CSV]"))
	}
	switch app.cursorRow.Term {
	case "\r\n":
		n += first(io.WriteString(app.out, "[CRLF]"))
	case "\n":
		n += first(io.WriteString(app.out, "[LF]"))
	case "":
		n += first(io.WriteString(app.out, "[EOF]"))
	}
	if app.Mode.HasBom() {
		n += first(io.WriteString(app.out, "[BOM]"))
	}
	if app.Mode.NonUTF8 {
		if app.Mode.IsUTF16LE() {
			n += first(io.WriteString(app.out, "[16LE]"))
		} else if app.Mode.IsUTF16BE() {
			n += first(io.WriteString(app.out, "[16BE]"))
		} else {
			n += first(io.WriteString(app.out, "[ANSI]"))
		}
	}
	if 0 <= app.cursorCol && app.cursorCol < len(app.cursorRow.Cell) {
		n += first(fmt.Fprintf(app.out, "(%d,%d/%d): ",
			app.cursorCol+1,
			app.cursorRow.lnum+1,
			app.cursorRow.list.Len()))
		var buffer strings.Builder
		buffer.WriteString(app.cursorRow.Cell[app.cursorCol].SourceText(app.Mode))
		if app.cursorCol < len(app.cursorRow.Cell)-1 {
			buffer.WriteByte(app.Mode.Comma)
		} else if term := app.cursorRow.Term; term != "" {
			buffer.WriteString(term)
		} else { // EOF
			buffer.WriteString("\u2592")
		}
		io.WriteString(app.out, runewidth.Truncate(replaceTable.Replace(buffer.String()), app.screenWidth-n, "..."))
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
	*application
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
	OutputSep       string
	SavePath        string
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

func (app *application) readlineAndValidate(prompt, text string, row *RowPtr, col int) (string, error) {
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

func (app *application) cmdQuit() (*Result, error) {
	if !app.ReadOnly && app.isDirty() {
		ch, err := app.MessageAndGetKey(`Quit: Save changes ? ["y": save, "n": quit without saving, other: cancel]`)
		if err != nil {
			return nil, err
		}
		if ch == "y" || ch == "Y" {
			message, err := app.cmdSave()
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(app.out, "\r%s%s%s", ansi.YELLOW, message, ansi.ERASE_LINE)
		} else if ch != "n" && ch != "N" {
			return nil, nil
		}
	}
	io.WriteString(app.out, "\n")
	return &Result{application: app}, nil

}

func (app *application) Fetch() (bool, *uncsv.Row, error) {
	if app.fetch == nil {
		return false, nil, nil
	}
	return app.fetch()
}

func (app *application) TryFetch() (bool, *uncsv.Row, error) {
	if app.tryFetch == nil {
		return false, nil, nil
	}
	return app.tryFetch()
}

func (app *application) nextOrFetch(p *RowPtr) *RowPtr {
	if next := p.Next(); next != nil {
		return next
	}
	if ok, row, _ := app.TryFetch(); ok {
		if row != nil {
			app.Push(row)
		}
		return p.Next()
	}
	return nil
}

func (cfg *Config) edit(fetch func() (*uncsv.Row, error), out io.Writer) (*Result, error) {
	if cfg.KeyMap == nil {
		cfg.KeyMap = make(map[string]func(*KeyEventArgs) (*CommandResult, error))
	}

	mode := cfg.Mode
	if mode == nil {
		mode = &uncsv.Mode{}
		cfg.Mode = mode
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
	app := cfg.newApplication(out)
	if fetch != nil {
		if row, err := fetch(); err == nil && !isEmptyRow(row) {
			app.Push(row)
		} else {
			newRow := uncsv.NewRow(mode)
			app.Push(&newRow)
		}
	}
	app.startRow = app.Front()
	app.startCol = 0
	app.cursorCol = 0
	app.cursorRow = app.Front()

	lastSearch := searchForward
	lastSearchRev := searchBackward
	lastWord := ""
	var lastWidth, lastHeight int

	keyWorker := nonblock.New(pilot.GetKey, func() (bool, *uncsv.Row, error) {
		if fetch == nil {
			return false, nil, nil
		}
		row, err := fetch()
		if err != nil {
			fetch = nil
			if !errors.Is(err, io.EOF) || isEmptyRow(row) {
				return false, nil, nil
			}
		}
		return true, row, err
	})
	defer keyWorker.Close()

	app.fetch = keyWorker.Fetch
	app.tryFetch = func() (bool, *uncsv.Row, error) {
		return keyWorker.TryFetch(100 * time.Millisecond)
	}

	var _screenHeight int
	var err error
	app.screenWidth, _screenHeight, err = pilot.Size()
	if err != nil {
		return nil, err
	}
	for _, title := range cfg.Titles {
		s, _ := cutStrInWidth(title, app.screenWidth-1)
		fmt.Fprintln(out, s)
	}
	message := cfg.Message
	var killbuffer pasteFunc
	for {
		app.screenHeight = _screenHeight - len(cfg.Titles)
		app.screenHeight -= cfg.HeaderLines
		if lastWidth != app.screenWidth || lastHeight != app.screenHeight {
			app.clearCache()
			lastWidth = app.screenWidth
			lastHeight = app.screenHeight
			io.WriteString(out, ansi.CURSOR_OFF)
		}

		app.Draw()

		io.WriteString(out, ansi.YELLOW)
		if message != "" {
			io.WriteString(out, runewidth.Truncate(message, app.screenWidth-1, ""))
		} else if 0 <= app.cursorRow.lnum && app.cursorRow.lnum < app.Len() {
			app.printStatusLine()
		}
		io.WriteString(out, ansi.RESET)
		io.WriteString(out, ansi.ERASE_SCRN_AFTER)

		const interval = 4
		displayUpdateTime := time.Now().Add(time.Second / interval)

		ch, err := keyWorker.GetOr(func(row *uncsv.Row, err error) bool {
			app.Push(row)
			if message == "" && (errors.Is(err, io.EOF) || time.Now().After(displayUpdateTime)) {
				io.WriteString(out, "\r"+ansi.YELLOW)
				app.printStatusLine()
				io.WriteString(out, ansi.RESET)
				io.WriteString(out, ansi.ERASE_SCRN_AFTER)
				displayUpdateTime = time.Now().Add(time.Second / interval)
			}
			return !errors.Is(err, io.EOF)
		})
		if err != nil {
			return nil, err
		}
		message = ""

		if handler, ok := cfg.KeyMap[ch]; ok {
			e := &KeyEventArgs{
				CursorRow:   app.cursorRow,
				CursorCol:   app.cursorCol,
				application: app,
			}
			cmdResult, err := handler(e)
			if err != nil || cmdResult.Quit {
				return &Result{application: app}, err
			}
			message = cmdResult.Message
		} else {
			switch ch {
			case keys.CtrlL:
				app.clearCache()
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
				for p := app.Front(); p != nil; p = p.Next() {
					p.Restore(mode)
				}
				app.resetSoftDirty()
				app.clearCache()
			case "q":
				if rc, err := app.cmdQuit(); err != nil {
					message = err.Error()
				} else if rc != nil {
					return rc, nil
				}
			case keys.CtrlF, keys.PageDown:
				for i := 0; i < app.screenHeight-1; i++ {
					if next := app.cursorRow.Next(); next != nil {
						app.cursorRow = next
					} else {
						break
					}
					if next := app.startRow.Next(); next != nil {
						app.startRow = next
					}
				}
			case keys.CtrlB, keys.PageUp:
				for i := 0; i < app.screenHeight-1; i++ {
					if prev := app.cursorRow.Prev(); prev != nil {
						app.cursorRow = prev
					} else {
						break
					}
					if prev := app.startRow.Prev(); prev != nil {
						app.startRow = prev
					}
				}
			case "j", keys.Down, keys.CtrlN, keys.Enter:
				if next := app.cursorRow.Next(); next != nil {
					app.cursorRow = next
				}
			case "k", keys.Up, keys.CtrlP:
				if prev := app.cursorRow.Prev(); prev != nil {
					app.cursorRow = prev
				}
			case "h", keys.Left, keys.ShiftTab:
				if app.cursorCol > 0 {
					app.cursorCol--
				}
			case "l", keys.Right, keys.CtrlI:
				app.cursorCol++
			case "0", "^", keys.CtrlA:
				app.cursorCol = 0
			case "$", keys.CtrlE:
				app.cursorCol = len(app.cursorRow.Cell) - 1
			case "g":
				ch, err := app.MessageAndGetKey("g- [\"g\": move to the beginning of file ]")
				if err == nil && ch == "g" {
					app.cursorRow = app.Front()
					app.startRow = app.Front()
					app.cursorCol = 0
					app.startCol = 0
				}
			case "<":
				app.cursorRow = app.Front()
				app.startRow = app.Front()
				app.cursorCol = 0
				app.startCol = 0
			case ">", "G":
				app.cursorRow = app.Back()
			case "n":
				if lastWord == "" {
					break
				}
				r, c := lastSearch(app.cursorRow, app.cursorCol, lastWord)
				if r == nil {
					message = fmt.Sprintf("%s: not found", lastWord)
					break
				}
				app.cursorRow = r
				app.cursorCol = c
			case "N":
				if lastWord == "" {
					break
				}
				r, c := lastSearchRev(app.cursorRow, app.cursorCol, lastWord)
				if r == nil {
					message = fmt.Sprintf("%s: not found", lastWord)
					break
				}
				app.cursorRow = r
				app.cursorCol = c
			case "*", "#":
				app.clearCache()
				if ch == "*" || ch == "#" {
					lastWord = app.cursorRow.Cell[app.cursorCol].Text()
				}
				if ch == "*" {
					lastSearch = searchExactForward
					lastSearchRev = searchExactBackward
				} else {
					lastSearch = searchExactBackward
					lastSearchRev = searchExactForward
				}
				r, c := lastSearch(app.cursorRow, app.cursorCol, lastWord)
				if r == nil {
					message = fmt.Sprintf("%s: not found", lastWord)
					break
				}
				app.cursorRow = r
				app.cursorCol = c
			case "/", "?":
				var err error
				app.clearCache()
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
				r, c := lastSearch(app.cursorRow, app.cursorCol, lastWord)
				if r == nil {
					message = fmt.Sprintf("%s: not found", lastWord)
					break
				}
				app.cursorRow = r
				app.cursorCol = c
			case "o":
				if cfg.ProtectHeader && app.cursorRow.lnum+1 < cfg.HeaderLines {
					message = msgProtectHeader
					break
				}
				if cfg.ReadOnly {
					message = msgReadOnly
					break
				}
				newRow := uncsv.NewRow(mode)
				newRow.Term = app.cursorRow.Term
				if app.cursorRow.Term == "" {
					app.cursorRow.Term = mode.DefaultTerm
				}
				if cfg.FixColumn {
					for len(newRow.Cell) < len(app.cursorRow.Cell) {
						newRow.Insert(0, "", mode)
					}
				}
				app.cursorRow = app.cursorRow.InsertAfter(&newRow)
				app.Repaint()
				app.clearCache()
				newCol := app.cursorCol
				if app.cursorCol >= len(app.cursorRow.Cell) {
					newCol = len(app.cursorRow.Cell) - 1
				}
				if text, err := app.readlineAndValidate("new line>", "", app.cursorRow, newCol); err == nil {
					app.cursorRow.Replace(newCol, text, mode)
				}
				app.setHardDirty()
			case "O":
				if m := cfg.checkWriteProtect(app.cursorRow); m != "" {
					message = m
					break
				}
				startPrevP := app.startRow.Prev()
				newRow := uncsv.NewRow(mode)
				if cfg.FixColumn {
					for len(newRow.Cell) < len(app.cursorRow.Cell) {
						newRow.Insert(0, "", mode)
					}
				}
				app.cursorRow = app.cursorRow.InsertBefore(&newRow)
				if startPrevP != nil {
					app.startRow = startPrevP.Next()
				} else {
					app.startRow = app.Front()
				}
				app.Repaint()
				app.clearCache()
				newCol := app.cursorCol
				if app.cursorCol >= len(app.cursorRow.Cell) {
					newCol = len(app.cursorRow.Cell) - 1
				}
				if text, err := app.readlineAndValidate("new line>", "", app.cursorRow, newCol); err == nil {
					app.cursorRow.Replace(newCol, text, mode)
				}
				app.setHardDirty()
			case "D":
				if m := cfg.checkWriteProtect(app.cursorRow); m != "" {
					message = m
					break
				}
				killbuffer = app.removeCurrentRow(&app.startRow, &app.cursorRow)
				app.Repaint()
				app.clearCache()
				app.setHardDirty()
			case "i":
				if m := cfg.checkWriteProtectAndColumn(app.cursorRow); m != "" {
					message = m
					break
				}
				app.clearCache()
				if text, err := app.readlineAndValidate("insert cell>", "", app.cursorRow, app.cursorCol); err == nil {
					if cells := app.cursorRow.Cell; len(cells) == 1 && cells[0].Text() == "" {
						app.cursorRow.Replace(app.cursorCol, text, mode)
					} else {
						app.cursorRow.Insert(app.cursorCol, text, mode)
						app.cursorCol++
					}
					app.setHardDirty()
				}
			case "a":
				if m := cfg.checkWriteProtectAndColumn(app.cursorRow); m != "" {
					message = m
					break
				}
				if cells := app.cursorRow.Cell; len(cells) == 1 && cells[0].Text() == "" {
					// current column is the last one and it is empty
					app.clearCache()
					if text, err := app.readlineAndValidate("append cell>", "", app.cursorRow, app.cursorCol); err == nil {
						app.cursorRow.Replace(app.cursorCol, text, mode)
					}
				} else {
					app.cursorCol++
					app.cursorRow.Insert(app.cursorCol, "", mode)
					app.Repaint()
					app.clearCache()
					if text, err := app.readlineAndValidate("append cell>", "", app.cursorRow, app.cursorCol); err != nil {
						// cancel
						app.cursorRow.Delete(app.cursorCol)
						app.cursorCol--
					} else {
						app.cursorRow.Replace(app.cursorCol, text, mode)
					}
				}
				app.setHardDirty()
			case "r", "R", keys.F2:
				if m := cfg.checkWriteProtect(app.cursorRow); m != "" {
					message = m
					break
				}
				cursor := &app.cursorRow.Cell[app.cursorCol]
				modifiedBefore := cursor.Modified()
				q := cursor.IsQuoted()
				app.clearCache()
				if text, err := app.readlineAndValidate("replace cell>", cursor.Text(), app.cursorRow, app.cursorCol); err == nil {
					app.cursorRow.Replace(app.cursorCol, text, mode)
					if q {
						*cursor = cursor.Quote(mode)
					}
				}
				modifiedAfter := cursor.Modified()
				app.updateSoftDirty(modifiedBefore, modifiedAfter)
			case "u":
				modifiedBefore := app.cursorRow.Cell[app.cursorCol].Modified()
				app.cursorRow.Cell[app.cursorCol].Restore(mode)
				app.updateSoftDirty(modifiedBefore, false)
			case "Y":
				killbuffer = app.yankCurrentRow(app.cursorRow)
			case "y":
				ch, err := app.MessageAndGetKey(`Yank ? ["l"/"v"/SPACE/TAB/C-F/→: cell, "y"/"r": row, "|"/"c": column]`)
				if err != nil {
					message = err.Error()
					break
				}
				switch ch {
				case "l", "v", " ", "\t", keys.CtrlF, keys.Right:
					killbuffer = app.yankCurrentCell(app.cursorRow, app.cursorCol)
				case "y", "r":
					killbuffer = app.yankCurrentRow(app.cursorRow)
				case "|", "c":
					killbuffer = app.yankCurrentColumn(app.cursorCol)
				}
			case "d":
				if m := cfg.checkWriteProtect(app.cursorRow); m != "" {
					message = m
					break
				}
				if cfg.FixColumn {
					ch, err = app.MessageAndGetKey(`Delete ? ["d"/"r": row]`)
				} else {
					ch, err = app.MessageAndGetKey(`Delete ? ["l"/"v"/SPACE/TAB/C-F/→: cell, "d"/"r": row, "|"/"c": column]`)
				}
				if err != nil {
					message = err.Error()
					break
				}
				switch ch {
				case "l", "v", " ", "\t", keys.CtrlF, keys.Right:
					if m := cfg.checkWriteProtectAndColumn(app.cursorRow); m != "" {
						message = m
						break
					}
					killbuffer = app.removeCurrentCell(app.cursorRow, app.cursorCol)
				case "d", "r":
					killbuffer = app.removeCurrentRow(&app.startRow, &app.cursorRow)
					app.Repaint()
					app.clearCache()
				case "|", "c":
					if m := cfg.checkWriteProtectAndColumn(app.cursorRow); m != "" {
						message = m
						break
					}
					killbuffer = app.removeCurrentColumn(app.cursorCol)
					app.Repaint()
					app.clearCache()
				}
				app.setHardDirty()
			case "p", "P", keys.AltP:
				if killbuffer == nil {
					break
				}
				pt := pasteAfter
				if ch == "P" {
					pt = pasteBefore
				} else if ch == keys.AltP {
					pt = pasteOver
				}
				if err := killbuffer(&app.startRow, &app.cursorRow, &app.cursorCol, pt); err != nil {
					message = err.Error()
					break
				}
				app.Repaint()
				app.clearCache()
				app.setHardDirty()
			case "x":
				if m := cfg.checkWriteProtect(app.cursorRow); m != "" {
					message = m
					break
				}
				cursor := &app.cursorRow.Cell[app.cursorCol]
				modifiedBefore := cursor.Modified()
				q := cursor.IsQuoted()
				app.cursorRow.Replace(app.cursorCol, "", mode)
				if q {
					*cursor = cursor.Quote(mode)
				}
				modifiedAfter := cursor.Modified()
				app.updateSoftDirty(modifiedBefore, modifiedAfter)
			case "\"":
				cursor := &app.cursorRow.Cell[app.cursorCol]
				modifiedBefore := cursor.Modified()
				if cursor.IsQuoted() {
					app.cursorRow.Replace(app.cursorCol, cursor.Text(), mode)
				} else {
					*cursor = cursor.Quote(mode)
				}
				modifiedAfter := cursor.Modified()
				app.updateSoftDirty(modifiedBefore, modifiedAfter)
			case "w":
				if msg, err := app.cmdSave(); err != nil {
					message = err.Error()
				} else {
					message = msg
				}
				app.clearCache()
			case "]":
				if w := cellWidth.Get(app.cursorCol); w < 40 {
					cellWidth.Set(app.cursorCol, w+1)
				}
				app.clearCache()
			case "[":
				if w := cellWidth.Get(app.cursorCol) - 1; w > 3 {
					cellWidth.Set(app.cursorCol, w)
				}
				app.clearCache()
			case keys.Escape:
				ch, err = app.MessageAndGetKey(`Esc- ["q": quit, "p": paste]`)
				if err != nil {
					message = err.Error()
					break
				}
				switch ch {
				case "q", "Q":
					if rc, err := app.cmdQuit(); err != nil {
						message = err.Error()
					} else if rc != nil {
						return rc, nil
					}
				case "p", "P":
					if killbuffer == nil {
						break
					}
					if err := killbuffer(&app.startRow, &app.cursorRow, &app.cursorCol, pasteOver); err != nil {
						message = err.Error()
						break
					}
					app.setHardDirty()
					app.Repaint()
					app.clearCache()
				}
			}
		}
		if L := len(app.cursorRow.Cell); L <= 0 {
			app.cursorCol = 0
		} else if app.cursorCol >= L {
			app.cursorCol = L - 1
		}
		if app.cursorRow.lnum < app.startRow.lnum {
			app.startRow = app.cursorRow.Clone()
		} else if app.cursorRow.lnum >= app.startRow.lnum+app.screenHeight-1 {
			goal := app.cursorRow.lnum - (app.screenHeight - 1) + 1
			for app.startRow = app.cursorRow.Clone(); app.startRow.lnum > goal; {
				app.startRow = app.startRow.Prev()
			}
		}
		if app.cursorCol < app.startCol {
			app.startCol = app.cursorCol
		} else {
			cellWidth := cfg.CellWidth
			if cellWidth == nil {
				cellWidth = NewCellWidth()
			}
			for {
				w := sum(cellWidth.Get, app.startCol, app.cursorCol+1)
				if w <= app.screenWidth {
					break
				}
				app.startCol++
			}
		}
		app.Rewind()
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
