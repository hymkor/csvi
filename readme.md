csview - simple csv viewer/editor
=================================

<img src="./csview.png" />

```
$ csview FILENAME(...)
```

or

```
$ csview < FILENAME
```

* Support OS:
    * Windows & Linux (tested on WSL)
* Charactor code:
    * Windows: UTF8 or Current Code Page
    * Linux: UTF8
* Key-binding:
    * Move Cursor: HJKL(like vi) , Ctrl-F/B/N/P(like Emacs) , arrow-key
    * Search
        * `/` (search forward)
        * `?` (search backward)
        * `n` (search next)
        * `N` (search next reverse)
    * Edit:
        * `i` (insert cell)
        * `a` (append cell)
        * `r` (replace cell)
        * `d` (delete cell)
        * `w` (write to file)
    * Quit: "q" or ESC
* Option
    * -c ... use Comma as field-separator (default when suffix is `.csv`)
    * -t ... use TAB as field-separator (default when suffix is not `.csv`)
