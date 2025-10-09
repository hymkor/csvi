"CSVI" - ターミナル用CSVエディタ
============================
[![GoDev](https://pkg.go.dev/badge/github.com/hymkor/csvi)](https://pkg.go.dev/github.com/hymkor/csvi)

[\<English\>](./README.md) / **\<Japanese\>**

( macOS でも動作するはずですが、検証環境が開発側にないため、実験的なサポート状態です →サポートバージョン: [v1.13.0](https://github.com/hymkor/csvi/releases/tag/v1.13.0) )
**CSVI** は、UNIX系システムや Windows のターミナル上で動作する CSV エディタです。

### &#10024; 主な特徴

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
* `-w widths` セルの幅を設定 (例: -w 14,0:10,1:20 1列目は10桁,2列目は20桁,他は14桁とする)
* `-fixcol` セルの挿入削除を禁止する (`i`, `a`, `x` を無効化)
* `-p` ヘッダー行を保護する
* `-readonly` 読み取り専用モード

[IANA名]: https://www.iana.org/assignments/character-sets/character-sets.xhtml

キーバインド
-----------

* カーソル移動
    * `h`,`Ctrl`-`B`,`←`,`Shift`-`TAB` (左)
    * `j`,`Ctrl`-`N`,`↓`,`Enter` (下)
    * `k`,`Ctrl`-`P`,`↑` (上)
    * `l`,`Ctrl`-`F`,`→`,`TAB` (右)
    * `<` (ファイル先頭)
    * `>`,`G` (ファイル末尾)
    * `0`,`^`,`Ctrl`-`A` (行頭)
    * `$`,`Ctrl`-`E` (行末)
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
    * `d`,`x` (現在のセルを削除)
    * `w` (ファイルもしくは標準出力(`'-'`)に出力する)
    * `o` (現在の行の後に新しい行を追加する)
    * `O` (現在の行の前に新しい行を挿入する)
    * `D` (現在の行を削除する)
    * `"` (可能であれば、二重引用符の囲む/外す)
    * `u` (現在のセルの元の値を復元する)
    * `y` (現在のセルの値を内部クリップボードへコピー)
    * `p` (現在のセルに内部クリップボードの値をペースト)
* `L` (指定したエンコーディングでファイルを再読み込み)
* 再表示: `Ctrl`-`L`
* 終了: `q` or `ESC`

Readline with SKK[^SKK]
-----------------------

環境変数 GOREADLINESKK に次のように辞書ファイルが指定されている時、[go-readline-skk] を使った内蔵SKKが使用できます

- Windows
    - `set GOREADLINESKK=SYSTEMJISYOPATH1;SYSTEMJISYOPATH2...;user=USERJISYOPATH`
    - (example) `set GOREADLINESKK=~/Share/Etc/SKK-JISYO.L;~/Share/Etc/SKK-JISYO.emoji;user=~/.go-skk-jisyo`
- Linux
    - `export GOREADLINE=SYSTEMJISYOPATH1:SYSTEMJISYOPATH2...:user=USERJISYOPATH`

(注: `~` はWindowsでも`cmd.exe`内であってもアプリ側で %USERPROFILE% へ自動で展開します)


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

Acknowledgements
----------------

- [sergeevabc (Aleksandr Sergeev)](https://github.com/sergeevabc) — [Issue #1](https://github.com/hymkor/csvi/issues/1)
- [kevin-gwyrdh (Kevin)](https://github.com/kevin-gwyrdh) — [Issue #4](https://github.com/hymkor/csvi/issues/4)
- [emisjerry (emisjerry)](https://github.com/emisjerry) — [YouTube動画](https://www.youtube.com/watch?v=_cxBQKpfUds)
- [rinodrops (Rino)](https://github.com/rinodrops) — [Discussion #5](https://github.com/hymkor/csvi/discussions/5#discussioncomment-13140997)

Author
------

- [hymkor (HAYAMA Kaoru)](https://github.com/hymkor)
