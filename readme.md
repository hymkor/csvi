csview - simple csv viewer
==========================

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
    * Search with `/`,`?`,`n` and `N`.
    * Quit: "q" or ESC
* Option
    * -c ... use Comma as field-separator (default when suffix is not `.tsv`)
    * -t ... use TAB as field-separator (default when suffix is `.tsv`)
