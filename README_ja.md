"CSVI" - ターミナル用CSVエディタ
============================

<!-- badges.cmd | -->
[![Go Test](https://github.com/hymkor/csvi/actions/workflows/go.yml/badge.svg)](https://github.com/hymkor/csvi/actions/workflows/go.yml)
[![License](https://img.shields.io/badge/License-MIT-red)](https://github.com/hymkor/csvi/blob/master/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/hymkor/csvi.svg)](https://pkg.go.dev/github.com/hymkor/csvi)
<!-- -->

( [\<English\>](./README.md) / **\<Japanese\>** )

**CSVI** は、UNIX系システムや Windows のターミナル上で動作する CSV エディタです。

## 主な特徴

- **保存時の差分が最小**  
  編集していないセルは、元のテキストの表現（改行コード、二重引用符、BOM、エンコーディング、区切り文字など）を極力維持します。  
  そのため、本当に加えた変更だけが差分として現れます。実データを安全に編集したい場合に最適です。

- **vi風のカーソル移動、Emacs風のセル編集**  
  `h/j/k/l` などで移動、`Ctrl` 系のキーで編集できます。

- **ファイル／標準入力の両方に対応**  
  CSVファイルを直接開くだけでなく、パイプ経由のデータも読み込めます。

- **高速な起動とバックグラウンド読込**  
  ファイルをすばやく開きつつ、読み込み処理は裏で進行します。

- **変更の視覚的な表示**  
  編集したセルには下線が表示され、`u`キーで変更前の状態に戻せます。

- **元データの構文情報を表示**  
  引用符の有無、区切り文字、文字コードなどの詳細を、画面最下行に表示します。

- **多様な文字コードのサポート**  
  - UTF-8（デフォルト）  
  - UTF-16  
  - Windows の現在のコードページ（自動検出）  
  - [IANA registry] に登録された任意のエンコーディング（`-iana NAME` で指定）

- **配色設定**
  - 既定では黒背景向けの配色です。
  - `-rv` オプションで白背景向け配色に切り替えられます。
  - 環境変数 `NO_COLOR` が定義されている場合はカラー表示を抑制します( https://no-color.org/ )

[IANA registry]: http://www.iana.org/assignments/character-sets/character-sets.xhtml

![image](./csvi.gif)

[@emisjerry](https://github.com/emisjerry) さんによる [紹介動画](https://www.youtube.com/watch?v=_cxBQKpfUds)

Install
-------

### Manual Installation

[Releases](https://github.com/hymkor/csvi/releases) よりバイナリパッケージをダウンロードして、実行ファイルを展開してください

> &#9888;&#65039; Note: macOS用バイナリは実験的ビルドで、検証できていません。
> もし何らかの問題を確認されましたらお知らせください！

### "go install" を使う場合

```
go install github.com/hymkor/csvi@latest
```

### scoop インストーラーを使う場合 (Windowsのみ)

```
scoop install https://raw.githubusercontent.com/hymkor/csvi/master/csvi.json
```

もしくは

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

* `-help` 本ヘルプを表示
* `-h int` ヘッダ行の行数
* `-c` 列区切りにカンマを使う(拡張子が `.csv` の時のデフォルト動作)
* `-t` 列区切りにタブを使う(拡張子が `.csv` でない時のデフォルト動作)
* `-semicolon` 区切りにセミコロンを使う
* `-iana string` 非UTF8テキストを読み書きする時の [IANA名] を指定する
* `-16be` UTF-16BE と判断する
* `-16le` UTF-16LE と判断する
* `-auto string` 自動処理 (テストコード用)
* `-nonutf8` UTF-8 と判断しない
* `-w widths` セルの幅を設定 (例: `-w 14,0:10,1:20` 1列目は10桁,2列目は20桁,他は14桁とする)
* `-fixcol` セルの挿入削除を禁止する (`i`, `a`, `x` を無効化)
* `-p` ヘッダー行を保護する
* `-readonly` 読み取り専用モード
* `-rv` 反転表示を有効にする（文字色と背景色を反転）

[IANA名]: https://www.iana.org/assignments/character-sets/character-sets.xhtml

キーバインド
-----------

* カーソル移動
    * `h`, `Ctrl`+`B`, `←`,`Shift`+`TAB` (左)
    * `j`, `Ctrl`+`N`, `↓`,`Enter` (下)
    * `k`, `Ctrl`+`P`, `↑` (上)
    * `l`, `Ctrl`+`F`, `→`,`TAB` (右)
    * `<` (ファイル先頭)
    * `>`, `G` (ファイル末尾)
    * `0`, `^`, `Ctrl`+`A` (行頭)
    * `$`, `Ctrl`+`E` (行末)
* 検索
    * `/` (キーワードを部分一致で前方検索)
    * `?` (キーワードを部分一致で後方検索)
    * `n` (次検索)
    * `N` (逆検索)
    * `*` (現在のセルの内容と完全一致する次のセルを検索)
    * `#` (現在のセルの内容と完全一致する前のセルを検索)
* 編集
    * `i` (現在のセルの前に新セルを挿入)
    * `a` (現在のセルの右に新セルを挿入)
    * `r` (現在のセルを置換)
    * `dl`, `d`+`SPACE`, `d`+`TAB`, `dv`, `x` (現在のセルを削除)
    * `dd`, `dr`, `D` (現在の行を削除する)
    * `dc`, `d|` (現在の列を削除する)
    * `w` (ファイルもしくは標準出力(`'-'`)に出力する)
    * `o` (現在の行の後に新しい行を追加する)
    * `O` (現在の行の前に新しい行を挿入する)
    * `"` (可能であれば、二重引用符の囲む/外す)
    * `u` (現在のセルの元の値を復元する)
    * `yl`, `y`+`SPACE`, `y`+`TAB`, `yv` (現在のセルを内部クリップボードへコピー)
    * `yy`, `yr`, `Y` (現在の行を内部クリップボードへコピー)
    * `yc`, `y|` (現在の列を内部クリップボードへコピー)
    * `p` (現在のセル/列/行の直後に内部クリップボードの値をペースト)
    * `P` (現在のセル/列/行の直前に内部クリップボードの値をペースト)
    * `ALT`+`p`, `ESC`+`p` (現在のセル/列/行を内部クリップボードの値で上書き)
* 表示設定
    * `L` (指定したエンコーディングでファイルを再読み込み)
    * `Ctrl`+`L` (再表示)
    * `]` (カーソルのある列の幅を広げる)
    * `[` (カーソルのある列の幅を縮める)
* 終了: `q` or `ESC`+`q`

環境変数
--------

### NO\_COLOR

環境変数 `NO_COLOR` が 1 文字以上設定されている場合、csvi の色付け出力を無効化します。これは [NO\_COLOR](https://no-color.org) で提唱されている標準仕様に従った挙動です。

### RUNEWIDTH\_EASTASIAN

Unicode で「曖昧幅」とされる文字の表示桁数を明示的に指定します。

- 2桁幅にする場合：`set RUNEWIDTH_EASTASIAN=1`
- 1桁幅にする場合：`set RUNEWIDTH_EASTASIAN=0`（`1` 以外の任意の1文字以上で可）

### COLORFGBG

`(FG);(BG)` 形式で色が定義されている場合、前景色の値が背景色より小さいとき、白背景を前提とした配色（`-rv` オプション相当）を自動的に使用します。

なお、csvi は通常、端末のデフォルト色を示すエスケープコード `ESC[39m` および `ESC[49m` を使用します。そのため、`(FG);(BG)` で指定された色そのものを直接採用するわけではありません。この設定は主に、灰色背景行の濃淡を白寄り・黒寄りのどちらにするかの判定に用いられます。

### GOREADLINESKK

環境変数 `GOREADLINESKK` に辞書ファイルを指定すると、[go-readline-skk] を利用した内蔵 SKK かな漢字変換[^SKK]が有効になります。

- **Windows**
  - `set GOREADLINESKK=SYSTEMJISYOPATH1;SYSTEMJISYOPATH2...;user=USERJISYOPATH`
  - 例:
    `set GOREADLINESKK=~/Share/Etc/SKK-JISYO.L;~/Share/Etc/SKK-JISYO.emoji;user=~/.go-skk-jisyo`
- **Linux**
  - `export GOREADLINESKK=SYSTEMJISYOPATH1:SYSTEMJISYOPATH2...:user=USERJISYOPATH`

（注）`~` は Windows の `cmd.exe` 上でもアプリ側で `%USERPROFILE%` に自動展開されます。

[^SKK]: Simple Kana to Kanji conversion program. One of the Japanese input method editors.

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

Acknowledgements
----------------

- [sergeevabc (Aleksandr Sergeev)](https://github.com/sergeevabc) — [Issue #1](https://github.com/hymkor/csvi/issues/1)
- [kevin-gwyrdh (Kevin)](https://github.com/kevin-gwyrdh) — [Issue #4](https://github.com/hymkor/csvi/issues/4)
- [emisjerry (emisjerry)](https://github.com/emisjerry) — [YouTube動画](https://www.youtube.com/watch?v=_cxBQKpfUds)
- [rinodrops (Rino)](https://github.com/rinodrops) — [Discussion #5](https://github.com/hymkor/csvi/discussions/5#discussioncomment-13140997)

Author
------

- [hymkor (HAYAMA Kaoru)](https://github.com/hymkor)
