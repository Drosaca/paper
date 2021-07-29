package main

import (
	"paper/fn"
	"paper/qr"
	"paper/stat"
	"fmt"
)


var SysArgs *fn.Args = &fn.SysArgs



func missingParameters() error {
	if SysArgs.Paper && SysArgs.Digital || !SysArgs.Paper && !SysArgs.Digital {
		return fmt.Errorf("incoherent flags")
	}
	if SysArgs.Input == "" ||
		(SysArgs.Digital && !stat.Exists(SysArgs.Input)) || (SysArgs.Paper && !stat.IsFile(SysArgs.Input)){
		return fmt.Errorf("bad input or output file %s %t %t", SysArgs.Input, stat.IsFile(SysArgs.Input), stat.Exists(SysArgs.Input))
	}
	return nil
}

func distribute(qrCode qr.Qr) error {
	if SysArgs.Paper {
		err := qrCode.CreateQr()
		if err != nil {
			return err
		}
	} else {
		err := qrCode.ReadQr()
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	fn.ParseArgs()
	err := missingParameters()
	if err != nil {
		fmt.Println(err)
		return
	}
	qrCode := qr.NewQr(SysArgs.Input, SysArgs.Output)
	err = distribute(qrCode)
	if err != nil {
		fmt.Println(err)
	}
}
