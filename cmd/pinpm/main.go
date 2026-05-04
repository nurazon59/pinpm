package main

import (
	"fmt"
	"os"

	"github.com/nurazon59/pinpm/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
