CSView - Unbreakable CSV Viewer/Editor
======================================

CSView is a tool for viewing and editing CSV files on the Linux and Windows terminal.  
Here are some key features:

- Non-user-modified cells retain their original values
    - For cells that do not contain line breaks or commas, double quotation marks are not added or deleted , and the current status is kept
    - Line breaks are preserved as much as possible, even if the code differs between LF and CRLF for each line
    - If the file begins with a Byte Order Mark (BOM), it is preserved and not removed.
    - Byte Order Mark is only added if there is a BOM at the start of the file.
- CSView supports the following encodings:
    - UTF8 (default)
    - Current codepage on Windows (automatically detected)
    - Encodings specified by the [IANA registry] (-iana NAME)

[IANA registry]: http://www.iana.org/assignments/character-sets/character-sets.xhtml

![image](./csview.png)

Install
-------

Download the binary package from [Releases](https://github.com/zetamatta/csview/releases)
and extract the executable.

Usage
-----

```
$ csview [-iana ENCODING] FILENAME(...)
```

or

```
$ cat FILENAME | csview [-iana ENCODING]
```

Options

* `-c` ... use Comma as field-separator (default when suffix is `.csv`)
* `-t` ... use TAB as field-separator (default when suffix is not `.csv`)
* `-iana NAME` ... IANA-registered-name to decode/encode NonUTF8 text

Key-binding
-----------

* Move Cursor
    * `h`,`Ctrl`-`B`,&#x2190; (move cursor left)
    * `j`,`Ctrl`-`N`,&#x2193; (move cursor down)
    * `k`,`Ctrl`-`P`,&#x2191; (move cursor up)
    * `l`,`Ctrl`-`F`,&#x2192; (move cursor right)
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
    * `w` (write to the **new** file or STDOUT(`'-'`).)
    * `o` (append a new line after the current one)
    * `O` (insert a new line before the current one)
    * `D` (delete the current line)
    * `"` (enclose or remove double quotations if possible)
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

Release Note
------------

- [English](./release_note_en.md)
- [Japanese](./release_note_ja.md)
