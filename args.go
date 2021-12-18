package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type (
	Args struct {
		ParseFile string
		ParseOnly bool
	}
)

func parseArgs() (a Args, err error) {
	cmdArgs := os.Args
	if len(cmdArgs) < 2 {
		err = errors.New("specify some go-file path")
		return
	}
	a.ParseFile = cmdArgs[len(cmdArgs)-1]
	var maxArgIndex = len(cmdArgs) - 1
	for i := 1; i < maxArgIndex; i++ {
		var val string
		var hasVal bool
		var arg = cmdArgs[i]
		if s := strings.SplitN(arg, "=", 2); len(s) > 1 {
			arg = s[0]
			val = s[1]
			hasVal = true
		}
		switch arg {
		case "-h", "--help":
			showHelpAndStop()
		case "-P", "--parse-only":
			if !hasVal && i+1 < maxArgIndex {
				i++
				val = cmdArgs[i]
			}
			a.ParseOnly = boolArg(val)
		default:
			err = fmt.Errorf("unknown argument: %s", arg)
			return
		}
	}
	return
}

func boolArg(s string) bool {
	if s == "0" {
		return false
	}
	if strings.EqualFold(s, "false") {
		return false
	}
	if strings.EqualFold(s, "f") {
		return false
	}
	if strings.EqualFold(s, "no") {
		return false
	}
	// empty value == `true` also
	return true
}

func showHelpAndStop() {
	os.Exit(0)
}
