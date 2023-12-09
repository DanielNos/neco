package main

import (
	"fmt"
	"os"

	color "github.com/fatih/color"
)

func info(message string) {
	color.Set(color.FgHiWhite)
	fmt.Print( "[INFO] ");
	
	fmt.Println(message)
}

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


func errorPos(file *string, line, char uint, message string) {
	color.Set(color.FgHiRed)
	fmt.Fprintf(os.Stderr, "[ERROR] ")
	
	color.Set(color.FgHiCyan)
	fmt.Fprintf(os.Stderr, "%s %d:%d ", *file, line, char)

	color.Set(color.FgHiWhite)
	fmt.Fprintf(os.Stderr, "%s\n", message)
}

func errorCodePos(codePos *CodePos, message string) {
	color.Set(color.FgHiRed)
	fmt.Fprintf(os.Stderr, "[ERROR] ")

	color.Set(color.FgHiCyan)
	fmt.Fprintf(os.Stderr, "%s %d:%d ", *codePos.file, codePos.startLine, codePos.startChar)
	
	color.Set(color.FgHiWhite)
	fmt.Fprintf(os.Stderr, "%s\n", message)
}

func fatal(error_code int, message string) {
	error(message)
	os.Exit(error_code)
}
