package startup

import (
	"flag"

	"github.com/hymkor/struct2flag"
)

type Flag struct {
	CellWidth     string `flag:"w,set the \x60widths\x60 of cells like '-w DefaultWidth,COL0:WIDTH0,COL1:WIDTH1,...'. COLn is the index starting from 0"`
	Header        uint   `flag:"h,the number of row-header"`
	Tsv           bool   `flag:"t,use TAB as field-separator"`
	Csv           bool   `flag:"c,use Comma as field-separator"`
	Semicolon     bool   `flag:"semicolon,use Semicolon as field-separator"`
	Iana          string `flag:"iana,IANA-registered-name to decode/encode NonUTF8 text(for example: Shift_JIS,EUC-JP... )"`
	NonUTF8       bool   `flag:"nonutf8,do not judge as utf8"`
	Help          bool   `flag:"help,this help"`
	Auto          string `flag:"auto,autopilot"`
	Utf16le       bool   `flag:"16le,Force read/write as UTF-16LE"`
	Utf16be       bool   `flag:"16be,Force read/write as UTF-16BE"`
	FixColumn     bool   `flag:"fixcol,Do not insert/delete a column"`
	ReadOnly      bool   `flag:"readonly,Read Only Mode"`
	ProtectHeader bool   `flag:"p,Protect the header line"`
	Title         string `flag:"title,Set title string"`
	ReverseVideo  bool   `flag:"rv,Enable reverse-video display (invert foreground and background colors)"`
	flagSet       *flag.FlagSet
}

func NewFlag() *Flag {
	return &Flag{
		CellWidth: "14",
		Header:    1,
	}
}

func (this *Flag) Bind(fs *flag.FlagSet) *Flag {
	this.flagSet = fs
	struct2flag.Bind(fs, this)
	return this
}

func Run() error {
	f := NewFlag().Bind(flag.CommandLine)
	flag.Parse()
	return f.Run()
}
