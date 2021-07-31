package fn

import (
	"flag"
	"fmt"
)

var SysArgs = Args{}

type Args struct {
	Output   string
	Input    string
	File     string
	Paper    bool
	Digital  bool
	Verbose  bool
	Password string
}

func ParseArgs() {
	flag.BoolVar(&SysArgs.Paper, "paper", false, " aka paperify -> creates the Qr code from ONE FILE")
	flag.BoolVar(&SysArgs.Digital, "digital", false, "aka digitalify -> reads the Qr code and writes THE FILE to output path")
	flag.BoolVar(&SysArgs.Verbose, "v", false, "verbose")
	flag.StringVar(&SysArgs.Output, "o", ".", "output path")
	flag.StringVar(&SysArgs.Input, "i", "", "input file path")
	flag.Parse()
}

func Log(a ...interface{})  {
	if SysArgs.Verbose {
		fmt.Println(a...)
	}
}
