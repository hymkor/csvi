- "a","o","O": make new cell and repaint before getline is called
- Readline: Ctrl-P: fetch the value of the cell above the same column
- Readline: TAB: complete with the values of the cell above the same column
- In principle, data other than cells changed by the user will remain as they are
    - If ByteOrderMark is attached to the beginning of the file, do not delete
    - Do not insert ByteOrderMark if there is no BOM at the beginning of the file
    - For cells that do not contain line breaks or commas, double quotation marks are not added or deleted , and the current status is kept
    - Even if the line break code is different from LF or CRLF for each line, maintain it as much as possible.
- "a": works same as "r" when the current line is empty
- "w": support filename completion
- Enabled to specify encoding other than UTF8 with `-iana NAME` (mainly for Linux)
- Cell source data is now displayed on the status line.
- Draw underline on the modified cells
- Implement `"`: enclose or remove double quotations if possible
- Implement `u`: restore the original value of the current cell
- Fix: cell width was incorrect when it contained characters whose widths are ambiguous

v1.1.3
======
Feb 16 2024

- Fix: the attributes of text converted by SKK were incorrect on Windows 8.1

v1.1.2
=====
Oct 01 2023

- Strings being converted with SKK are now displayed as reversed or underlined
- Fix: SKK failed to start when user-jisyo file did not exist

v1.1.1
======
Sep 20 2023

- Use `:` for the path list separator instead of `;` from %GOREADLINESKK% on Linux

v1.1.0
======
Sep 20 2023

- Backport from [lispread]
    - Implement 'y'(yank) and 'p'(paste)
    - "o" and "O" query the text for the new cell now
    - Fix: error was not reported when the specified file is a directory
    - When no arguments are given and stdin is terminal, start with 1 cell immidiately
    - Support [go-readline-skk]

[lispread]: https://github.com/hymkor/lispread
[go-readline-skk]: https://github.com/nyaosorg/go-readline-skk

v1.0.0
======
Sep 11 2023

- Fix for the the imcompatibility between v0.8.3 and v0.14.0 of go-readline-ny

v0.6.2
======
Nov 23 2022

- Fix: (#3) Too long field breaks the screen layout

v0.6.1
======
Feb 19 2022

- Display [TSV],[CSV],[LF],[CRLF] on the status line.

v0.6.0
======
Dec 10 2021

- Change visual:
    - Change the field width 12 to 14
    - Change the background pattern: blue-ichimatsu -> gray-stripe
    - Show all cell string when the rightside cell is empty
    - Show `[BOM]``[ANSI]` marks
- `w` can override exist file
    - Output with ansi-encoding if input file is encoded by ansi-encoding
    - Fix: on Linux, the size of the output was zero bytes
    - BOM is restored to the saved file when original file has a BOM
- Fix: empty lines in the input data were ignored.
- `x`: assign delete cell same as `d`

v0.5.0
======
Mar 27 2020

- `o` - append a new line after the current line
- `O` - insert a new line before the current line
- `D` - delete the current line

v0.4.0
======
Nov 4 2019

- Support window resized
- Implement Ctrl-L repaint
- `w`: (save)
    - field separator for output becomes one for input now
    - do not overwrite to a existing file
    - default fname is args[0] or "-"
    - filename '-' means stdout
- Use stderr for drawing rather than stdout
- `q`: (quit) ask yes/no

v0.3.0
======
Nov 2 2019

- Support editing and writing to the file.

v0.2.0
======
Oct 31 2019

- Implement search command `/`,`?`,`n`,`N`

v0.1.0
======
Oct 27 2019

- first release
