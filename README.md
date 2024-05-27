"CSVI" - Terminal CSV Editor
============================
[![GoDev](https://pkg.go.dev/badge/github.com/hymkor/csvi)](https://pkg.go.dev/github.com/hymkor/csvi)

**&lt;English&gt;** / [&lt;Japanese&gt;](./README_ja.md)

- *Since the version 1.6.0, CSView is renamed to CSVI because not a few products that have the same name exist in the same category.*

CSVI is the CSV editor that runs on the terminal of Linux and Windows.
Here are some key features:

- Keybinding: vi-like on moving cursor and Emacs-like on editing cell
- It reads the data from both file and standard-input
- Start quickly and load data in the background
- Modified cells are displayed with underline
    - With one key `u`, original value before modifying can be restored
- Non-user-modified cells retain their original values
    - Enclosing double quotations or not of the cell value that contains neither commas nor line breaks
    - LF or CRLF for line breaks
    - BOM of the beginning of files
    - The representation before decoding double quotations, encoding, field and record sperators and so on are displayed on the bottom line
- CSVI supports the following encodings:
    - UTF8 (default)
    - UTF16
    - Current codepage on Windows (automatically detected)
    - Encodings specified by the [IANA registry] (-iana NAME)

[IANA registry]: http://www.iana.org/assignments/character-sets/character-sets.xhtml

![image](./csvi.gif)

[Video](https://www.youtube.com/watch?v=_cxBQKpfUds) by [@emisjerry](https://github.com/emisjerry)

Install
-------

### Manual Installation

Download the binary package from [Releases](https://github.com/hymkor/csvi/releases) and extract the executable.

### Use "go install"

```
go install github.com/hymkor/csvi/cmd/csvi@latest
```

### Use scoop-installer

```
scoop install https://raw.githubusercontent.com/hymkor/csvi/master/csvi.json
```

or

```
scoop bucket add hymkor https://github.com/hymkor/scoop-bucket
scoop install csvi
```

Usage
-----

```
$ csvi {options} FILENAME(...)
```

or

```
$ cat FILENAME | csvi {options}
```

Options

* `-help` this help
* `-h int` the number of fixed header lines
* `-c` use Comma as field-separator (default when suffix is `.csv`)
* `-t` use TAB as field-separator (default when suffix is not `.csv`)
* `-semicolon` use Semicolon as field-separator
* `-iana string` [IANA-registered-name] to decode/encode NonUTF8 text
* `-16be` Force read/write as UTF-16BE
* `-16le` Force read/write as UTF-16LE
* `-auto string` auto pilot (for testcode)
* `-nonutf8` do not judge as UTF-8
* `-w uint` set the width of cell (default 14)

[IANA-registered-name]: https://www.iana.org/assignments/character-sets/character-sets.xhtml

Key-binding
-----------

* Move Cursor
    * `h`,`Ctrl`-`B`,`←`,`Shift`-`TAB` (move cursor left)
    * `j`,`Ctrl`-`N`,`↓`,`Enter` (move cursor down)
    * `k`,`Ctrl`-`P`,`↑` (move cursor up)
    * `l`,`Ctrl`-`F`,`←`,`TAB` (move cursor right)
    * `<` (move the beginning of file)
    * `>`,`G` (move the end of file)
    * `0`,`^`,`Ctrl`-`A` (move the beginning of the current line)
    * `$`,`Ctrl`-`E` (move the end of the current line)
* Search
    * `/` (search forward)
    * `?` (search backward)
    * `n` (search next)
    * `N` (search next reverse)
* Edit
    * `i` (insert a new cell before the current one)
    * `a` (append a new cell after the current one)
    * `r` (replace the current cell)
    * `d`,`x` (delete the current cell)
    * `w` (write to a file or STDOUT(`'-'`))
    * `o` (append a new line after the current one)
    * `O` (insert a new line before the current one)
    * `D` (delete the current line)
    * `"` (enclose or remove double quotations if possible)
    * `u` (restore the original value of the current cell)
    * `y` (copy the value of the current cell to kill-buffer)
    * `p` (paste the value of kill-buffer to the current cell)
* Repaint: `Ctrl`-`L`
* Quit: `q` or `ESC`

Readline with SKK[^SKK]
-----------------------

When the environment variable GOREADLINESKK is defined, [go-readline-skk] is used.

- Windows
    - `set GOREADLINESKK=SYSTEMJISYOPATH1;SYSTEMJISYOPATH2...;user=USERJISYOPATH`
    - (example) `set GOREADLINESKK=~/Share/Etc/SKK-JISYO.L;~/Share/Etc/SKK-JISYO.emoji;user=~/.go-skk-jisyo`
- Linux
    - `export GOREADLINE=SYSTEMJISYOPATH1:SYSTEMJISYOPATH2...:user=USERJISYOPATH`

[^SKK]: Simple Kana to Kanji conversion program. One of the Japanese input method editor.

[go-readline-skk]: https://github.com/nyaosorg/go-readline-skk

Use as a package
----------------

```example.go
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/mattn/go-colorable"

    "github.com/hymkor/csvi"
    "github.com/hymkor/csvi/uncsv"
)

func main() {
    source := `A,B,C,D
"A1","B1","C1","D1"
"A2","B2","C2","D2"`

    cfg := &csvi.Config{
        Mode: &uncsv.Mode{Comma: ','},
    }

    result, err := cfg.Edit(strings.NewReader(source), colorable.NewColorableStdout())

    if err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        os.Exit(1)
    }

    // // env GOEXPERIMENT=rangefunc go run example
    // for row := range result.Each {
    //     os.Stdout.Write(row.Rebuild(cfg.Mode))
    // }
    result.Each(func(row *uncsv.Row) bool {
        os.Stdout.Write(row.Rebuild(cfg.Mode))
        return true
    })
}
```

Release Note
------------

- [English](./release_note_en.md)
- [Japanese](./release_note_ja.md)
