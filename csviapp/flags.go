package csviapp

import (
	"flag"
	"path/filepath"

	"github.com/hymkor/struct2flag"
)

type Options struct {
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
	OutputSep     string `flag:"ofs,Output separator between cells"`
	SavePath      string
	flagSet       *flag.FlagSet
}

func NewOptions() *Options {
	return &Options{
		CellWidth: "14",
		Header:    1,
	}
}

func (f *Options) Bind(fs *flag.FlagSet) *Options {
	f.flagSet = fs
	struct2flag.Bind(fs, f)
	return f
}

func Run() error {
	f := NewOptions().Bind(flag.CommandLine)
	flag.Parse()

	if args := flag.Args(); len(args) >= 1 {
		var err error
		f.SavePath, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}

	return f.Run()
}
