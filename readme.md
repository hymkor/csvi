CSView
======

CSView is the simple CSV viewer/editor running on the terminal of the Linux and Windows.  
It supports UTF8 and the Windows current codepage(ANSI)

![image](./csview.png)

### Install

Download the binary package from [Releases](https://github.com/zetamatta/csview/releases)
and extract the executable.

### Usage

```
$ csview FILENAME(...)
```

or

```
$ cat FILENAME | csview
```

Options

* `-c` ... use Comma as field-separator (default when suffix is `.csv`)
* `-t` ... use TAB as field-separator (default when suffix is not `.csv`)

### Key-binding:

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
    * `d` (delete the current cell)
    * `w` (write to the **new** file or STDOUT(`'-'`).)
    * `o` (append a new line after the current one)
    * `O` (insert a new line before the current one)
    * `D` (delete the current line)
* Repaint: `Ctrl`-`L`
* Quit: `q` or `ESC`

Release Note
============

- [English](./release_note_en.md)
- [Japanese](./release_note_ja.md)
