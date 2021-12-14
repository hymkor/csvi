
- Display [TSV] or [CSV] on the status line.

v0.6.0
------
on Dec.10, 2021

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
------
on Mar.27, 2020

- `o` - append a new line after the current line
- `O` - insert a new line before the current line
- `D` - delete the current line

v0.4.0
------
on Nov.4, 2019

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
------
on Nov.2, 2019

- Support editing and writing to the file.

v0.2.0
------
on Oct.31, 2019

- Implement search command `/`,`?`,`n`,`N`

v0.1.0
------
on Oct.27, 2019

- first release
