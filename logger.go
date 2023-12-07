package main

import (
	"fmt"
	"os"

	color "github.com/fatih/color"
)

func warning(message string) {
	color.Set(color.FgHiYellow)
	fmt.Print( "[WARNING] ");
	color.Set(color.FgHiWhite)

	fmt.Println(message)
}

func error(message string) {
	color.Set(color.FgHiRed)
	fmt.Fprint(os.Stderr, "[ERROR] ");
	color.Set(color.FgHiWhite)

	fmt.Fprintln(os.Stderr, message)
}

func fatal(error_code int, message string) {
	error(message)
	os.Exit(error_code)
}
