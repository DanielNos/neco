package main

import (
	"fmt"
	"os"
)

func fatal(error_code int, message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(error_code)
}
