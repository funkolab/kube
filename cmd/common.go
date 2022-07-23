package cmd

import "os"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func isInputFromPipe() bool {
	fileInfo, err := os.Stdin.Stat()
	check(err)
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
