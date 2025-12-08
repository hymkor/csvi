Release notes
=============
( **English** / [Japanese](release_note_ja.md) )

v1.17.1
-------
Dec 9, 2025

- Fixed an issue where `u` (undo) restored the value from the initial load instead of the last save.

v1.17.0
-------
Dec 7, 2025

### Specification Changes

- Truncate text that exceeds the cell width and append U+2026 (ellipsis). (#15 and #19 thanks to @toolleeo)
- Implement `-ofs STRING` to specify the separator between cells. (#15 and #19, thanks to @toolleeo)
- Map x to clear the current cell (#16, #21, thanks to @toolleeo)
  ( The previous behavior—deleting the cell and shifting the cells on the right—is still available via `dl`, `d`+`SPACE`, `d`+`TAB`, and `dv`. )
- Track whether the file has been modified (#23)
  - Display `*` in the status line when there are unsaved changes
  - When executing the `q` command, prompt whether to save the changes (#17, thanks to @toolleeo)
- Implement page up/down (#22, #24, thanks to @wumfi)
- Assign `gg` as an additional keybinding for moving to the first row. (#18, #25, thanks to @toolleeo)

### Internal Changes

- Replace all PowerShell-based test scripts with Go test files so that the test suite can run on GitHub Actions and other non-Windows environments. (#14)

v1.16.0
-------
Nov 24, 2025

### Bug Fixes

- Fix: Kill buffer of readline was cleared between sessions (#11)

### Specification Changes

- Add key bindings: (#7, #8 and #12, Thanks to @toolleeo)
    - `dl`, `d`+`SPACE`, `d`+`TAB`, `dv`, `x` (delete the current cell)
    - `dd`, `dr`, `D` (delete the current line)
    - `dc`, `d|` (delete the current column)
    - `yl`, `y`+`SPACE`, `y`+`TAB`, `yv` (copy the values of the current cell to kill-buffer)
    - `yy`, `yr`, `Y` (copy the values of the current row to kill-buffer)
    - `yc`, `y|` (copy the values of the current column to kill-buffer)
    - `p` (paste the values of kill-buffer after the current cell, row or column)
    - `P` (paste the values of kill-buffer before the current cell, row or column)
    - `ALT`-`p`, `ESC`+`p` (overwrite the current cell/row/column with the content of the kill-buffer)
- Assigned `ESC`+`q` as quitting because `ESC` had been assigned for quitting before, but is now used as a prefix key for 2-stroke commands (`q` still works as before). (#12)

### Internal Changes

- Internal maintenance: updated dependencies and removed staticcheck warnings. (#6)

v1.15.1
-------
Oct 19, 2025

### Specification Changes

- **Removed runtime measurement for ambiguous-width Unicode characters**
  Previously, `csvi` determined the display width of ambiguous-width characters (e.g., `∇`) by printing them and reading the cursor position using `ESC[6n]`.
  This caused issues on some older terminals and added unnecessary complexity, so the feature has been removed.
  The program now relies solely on [mattn/go-runewidth] for width determination.
  In most cases it works correctly, but if the width is misdetected, you can control it with the environment variable `RUNEWIDTH_EASTASIAN`:

  - Double-width: `set RUNEWIDTH_EASTASIAN=1`
  - Single-width: `set RUNEWIDTH_EASTASIAN=0` (any non-`1` value with at least one character is also valid)

  The options `-aw`, `-an`, and `-debug-bell` have been removed accordingly.

  [mattn/go-runewidth]: https://github.com/mattn/go-runewidth

- **Added automatic light-background detection**
  When the environment variable `COLORFGBG` is defined as `(FG);(BG)` and the foreground value `(FG)` is smaller than `(BG)`,
  `csvi` now automatically uses color settings for light backgrounds (equivalent to `-rv`).

### Bug Fixes

- Fixed an issue where executing

  ```
  echo "ihihi" | csvi -auto "w|-|q|y" > file
  ```

  resulted in unwanted text

  ```
  \r   Calibrating terminal... (press any key to skip)\r
  ```

  appearing at the beginning of the output file.
  This was caused by fallback handling in the old width-measurement logic, which has now been removed, so the problem no longer occurs.

### Internal Changes

- Moved command-line option parsing from the main package `cmd/csvi` to the subpackage `startup`.
- Removed the deprecated function `(Config) Main`.

v1.15.0
-------
Oct 10, 2025

- Added key bindings `]` and `[` to adjust the width of the current column (widen and narrow, respectively).
- Added `-rv` option to prevent unnatural colors on terminals with a white background
- At startup, the width of ambiguous-width Unicode characters was being measured, but on terminals that do not support the cursor position query sequence `ESC[6n`, this could cause a hang followed by an error. To address this:
    - `-aw` (treat ambiguous-width characters as 2 cells) and `-an` (treat ambiguous-width characters as 1 cell) options were added to skip the measurement and explicitly specify the character width.
    - If `ESC[6n` is not supported, the program now continues without aborting.
- Suppress color output if the `NO_COLOR` environment variable is set (following https://no-color.org/ )
- Added support for FreeBSD/amd64
- Added API functions `(Config) EditFromStringSlice`, `uncsv.NewRowFromStringSlice` and and `(*_Application) MessageAndGetKey`.
- Split the `"csvi"` package into subpackages such as `"internal/ansi"`, `"internal/manualctl"`, `"legacy"`, and `"candidate"`.

v1.14.0
-------
Jun 2, 2025

- Added `L` (Shift-L) command to reload the file using a specified encoding to correct detection errors
   - Encoding name completion is supported.
   - UTF-16LE and UTF-16BE are not supported yet.
- Added search command (`*` and `#`) to find the next occurrence of the current cell's content

v1.13.1
-------
Jun 25, 2025

- Made it possible to build with Go 1.20.14 to support Windows 7, 8, and Server 2008 or later.

v1.13.0
-------
May 14, 2025

### Experimental Release for macOS Users

This release includes a **macOS build for the first time**!

We don't have access to a macOS environment, so we can't test it ourselves.
If you're a macOS user and willing to help, **please try it and report back** through GitHub Issues or Discussions.

Thank you for your support!

### Other changes

- Added `-title` option to display a title row above the header row.

v1.12.0
-------
Oct 8, 2024

- [#4] Define the different width for cells. e.g., `-w 14,0:10,1:20` means: the first-column has 10 characters wide, the second 20, and other 14. (Thanks to [@kevin-gwyrdh])
- Fix: panic when 0 bytes files (`nul` or `/dev/null`) were given

[#4]: https://github.com/hymkor/csvi/issues/4
[@kevin-gwyrdh]: https://github.com/kevin-gwyrdh

v1.11.1
-------
Oct 7, 2024

- Fix: suggestions did not start when the original value of the current cell was empty
- Fix: the search target for suggestions shifted one column to the right when inserting a cell value with `a`.

v1.11.0
-------
Oct 6, 2024

- While entering text in a cell, automatically display suggestions from cells in the same column that contain the current input. Press → or Ctrl-F to accept. [go-readline-ny v1.5.0]
- Support the hankaku-kana mode on the SKK input (`Ctrl-Q` to enter the hankaku-kana mode from SKK kana mode) [go-readline-skk v0.4.0]

[go-readline-ny v1.5.0]: https://github.com/nyaosorg/go-readline-ny/releases/tag/v1.5.0
[go-readline-skk v0.4.0]: https://github.com/nyaosorg/go-readline-skk/releases/tag/v0.4.0

v1.10.1
-------
Jun 10, 2024

- Modifying package
    - When the cell validation fails, prompt to modify the input text

v1.10.0
-------
Jun 02, 2024

- When `-fixcol` is specified
    - Fix: `o` and `O`: inserted column was always the first one of the new line
    - Fix: `O`: the line of cursor is incorrect before new cell text is input
- Add a new option to protect header (`-p` and `Config.ProtectHeader`)
- Do not create a row contains nothing but EOF.
- Modifying package
    - Added a mechanism for cell input validation
    - Change the parameter type of hander function for key pressed from Application to KeyEventArgs (Compatiblity broken)
    - Unexport type `Application` and `(*Config) Edit` returns `*Result` instead

v1.9.5
------
May 27, 2024

- Modifying package
    - User functions can be assigned to keys
    - `csvi.Result` is the alias of `csvi.Application` now

v1.9.4
------
May 26, 2024

- Fix: panic occured when no input lines were given. It is a bug existing only on v1.9.3 whose executable was not released
- Modifing package
    - Make `uncsv.Cell.Original()` that returns the original value before modified.
    - When `Config.FixColumn` is set true, the new row which `o` & `O` insert has all columns same as the row cursor exists
    - `csvi.Result` has removed rows in a field.

v1.9.3
------
May.17, 2024

- Modifing pacakge
    - Change the return value of `Config.Edit` from `(*RowPtr,error)` to `(*Result,error)`

v1.9.2
------
May.12, 2024

- Modifing package
    - Use 14 for the default of csvi.Config.CellWidth
    - Implement csvi.Config.Edit as a function instead of csvi.Config.Main

v1.9.1
------
May.09, 2024

- Fix timing to close the terminal input was incorrect
  (For some reason it hasn't surfaced as a problem)

v1.9.0
------
May.08, 2024

- Add the option `-fixcol` that disables keys `i`,`a`, and `x` not to shift columns.
- Move the main function to the sub-package `cmd/csvi` to be available as a package of Go
- Add the option `-readonly` that forbide changing the value of cell. When enabled, "q" shutdowns csvi immediately

v1.8.1
------
Apr.26, 2024

- Fix: crashed on starting `csvi` with no arguments
- Fix: a cell were not flipped when the cursor was in a cell with no text.
- Fix: the foreground color was not black, but gray.
- Change the case STDOUT or STDERR is used on no arguments to make the content of foo.txt becomes `foo\r\n` when executing `echo "foo" | csvi -auto "w|-|q|y" > foo.txt`

v1.8.0
------
Apr.24, 2024

- Update the read bytes of the status line 4 times per second.
- Reduced the number of times ERASELINE(ESC[K) is output for too slow terminal to improve the speed to update screen.

v1.7.1
------
Apr 16 2024

- Set cursor on or off when yes or no is asked.
- Fix the problem (since v.1.6.0) that the cursor position could become invalid after moving from a long line to a short line, causing a crash when editing.

v1.7.0
------
Apr 15 2024

- Added the `-auto` option to enable running automated tests even without Expect-Lua. Using this option, all test programs were rewritten in PowerShell. nkf32 is no longer required for testing.
- Added the `-16le` and `-16be` options to force interpretation as UTF-16 little-endian or big-endian encoding, respectively.
- `-semicolon`: Enabled using semicolons as field delimiters (for some European locales that use commas as decimal separators). Considered allowing arbitrary delimiter strings, but decided against it to avoid potential issues.
- `-nonutf8`: Added an option to handle cases where data is incorrectly interpreted as UTF-8 when it is not actually encoded that way.
- Added the `-help` option to display a list of available options.
- Increased the number of leading bytes checked to detect UTF-16 encoding from the previous value to 10 bytes.

v1.6.0
------
Apr 8 2024

- Rename from CSView to CSVI because not a few products that have the same name exist in the same category.
- Previously, users must have waited until all the lines were read, but now users can operate when the first 100 lines are read. The rest lines are read while waiting for key input.
- Improve memory efficiency by holding row data with "container/list" now, those were held with slice.
- Fix: `o` after `>`: the last line was joined with the previous line in the saved file.
- Prevent the displayed position from being incorrect even when it contains the character whose width is difficult to judge
- Fix: the problem abortion at starting on Windows 8.1
- Enable to build without `env GOEXPERIMENT=rangefunc`
- Show (CURRENT-COLUMN-POSITION,CURRENT-ROW-POSITION/ALL-READ-ROWS-NUMBER) on the status line.

v1.5.0
------
Mar 31 2024

- Support UTF16
    - Judge the file encoding UTF16 when the first two bytes are `\xFE\xFF` or `\xFF\xFE`, or `\0` is one of the two bytes

v1.4.0
------
Mar 27 2024

- Fix the problem the cache buffer for drawing did not work on v1.3.0
- Set 1 as the default value of `-h` option and the first line is fixed header on default
    - To disable, use `-h 0`
    - Change the type `uint` (unsigned integer)
- The width of the cells can be changed with `-w uint`
- Even when the double quotations get redundant as the result to edit, they are not removed now
- When inputting the save filename, the initial position of the cursor is now before the extension
- Modifying the package `uncsv`
    - Rename: `(Cell) ReadableSource` → `(Cell) SourceText()`
    - Implment: `(Cell) Source` that returns the binary value before decoding

v1.3.0
------
Mar 25 2024

- `[CRLF]` or `[LF]` in the status line now indicates the line feed code of the current line instead of the representative line feed code of the entire file.
- Rename sub-package: `csv`(`unbreakable-csv`) to `uncsv`(`uncsv`)
- The first few lines can now be fixed as header lines.(`-h int`)

v1.2.0
------
Feb 29 2024

- `a`,`o`,`O`: make new cell and repaint before getline is called
- Readline: Ctrl-P: fetch the value of the cell above the same column
- Readline: TAB: complete with the values of the cell above the same column
- In principle, data other than cells changed by the user will remain as they are
    - If ByteOrderMark is attached to the beginning of the file, do not delete
    - Do not insert ByteOrderMark if there is no BOM at the beginning of the file
    - For cells that do not contain line breaks or commas, double quotation marks are not added or deleted , and the current status is kept
    - Even if the line break code is different from LF or CRLF for each line, maintain it as much as possible.
- `a`: works same as `r` when the current line is empty
- `w`: support filename completion
- Enabled to specify encoding other than UTF8 with `-iana NAME` (mainly for Linux)
- Cell source data is now displayed on the status line.
- Draw underline on the modified cells
- Implement `"`: enclose or remove double quotations if possible
- Implement `u`: restore the original value of the current cell
- Fix: cell width was incorrect when it contained characters whose widths are ambiguous
- Add key assigns: `G`:Go to EOF, `Enter`:go to next line, `TAB`:go to the rightside cell, `Shift`+`TAB`:go to the leftside cell

v1.1.3
------
Feb 16 2024

- Fix: the attributes of text converted by SKK were incorrect on Windows 8.1

v1.1.2
------
Oct 01 2023

- Strings being converted with SKK are now displayed as reversed or underlined
- Fix: SKK failed to start when user-jisyo file did not exist

v1.1.1
------
Sep 20 2023

- Use `:` for the path list separator instead of `;` from %GOREADLINESKK% on Linux

v1.1.0
------
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
------
Sep 11 2023

- Fix for the the imcompatibility between v0.8.3 and v0.14.0 of go-readline-ny

v0.6.2
------
Nov 23 2022

- Fix: (#3) Too long field breaks the screen layout

v0.6.1
------
Feb 19 2022

- Display [TSV],[CSV],[LF],[CRLF] on the status line.

v0.6.0
------
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
------
Mar 27 2020

- `o` - append a new line after the current line
- `O` - insert a new line before the current line
- `D` - delete the current line

v0.4.0
------
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
------
Nov 2 2019

- Support editing and writing to the file.

v0.2.0
------
Oct 31 2019

- Implement search command `/`,`?`,`n`,`N`

v0.1.0
------
Oct 27 2019

- first release
