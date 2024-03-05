package logger

import (
	"bufio"
	"fmt"
	data "neco/dataStructures"
	"os"
	"strings"

	color "github.com/fatih/color"
)

const (
	LL_Info byte = iota
	LL_Warning
	LL_Error
)

var LoggingLevel byte = LL_Info

func readLine(filePath string, lineIndex uint) (string, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	// Create scanner
	scanner := bufio.NewScanner(file)
	var currentLine uint = 0

	// Read lines until correct line number
	for currentLine < lineIndex && scanner.Scan() {
		currentLine++
	}

	// Scanner stopped with an error
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %s", err)
	}

	return scanner.Text(), nil
}

func Success(message string) {
	if LoggingLevel > LL_Info {
		return
	}

	color.Set(color.FgHiGreen)
	fmt.Print("[SUCCESS] ")
	color.Set(color.FgHiWhite)

	fmt.Println(message)
}

func Info(message string) {
	if LoggingLevel > LL_Info {
		return
	}

	color.Set(color.FgHiWhite)
	fmt.Print("[INFO]    ")

	fmt.Println(message)
}

func Warning(message string) {
	if LoggingLevel > LL_Warning {
		return
	}

	color.Set(color.FgHiYellow)
	fmt.Print("[WARNING] ")
	color.Set(color.FgHiWhite)

	fmt.Println(message)
}

func Error(message string) {
	color.Set(color.FgHiRed)
	fmt.Fprint(os.Stderr, "[ERROR]   ")
	color.Set(color.FgHiWhite)

	fmt.Fprintln(os.Stderr, message)
}

func ErrorPos(file *string, line, startChar, endChar uint, message string) {
	// Read line
	lineString, err := readLine(*file, line)

	if int(endChar) > len(lineString) {
		endChar--
	}

	// Print line
	if err == nil && len(strings.Trim(lineString, "\n \t")) != 0 {
		color.Set(color.FgWhite)

		fmt.Fprint(os.Stderr, lineString[0:startChar-1])

		color.Set(color.FgHiRed)
		color.Set(color.Underline)

		fmt.Fprint(os.Stderr, lineString[startChar-1:endChar])

		color.Set(color.Reset)
		color.Set(color.FgWhite)

		fmt.Fprint(os.Stderr, lineString[endChar:])
		fmt.Fprintln(os.Stderr, "\n")
	}

	// Print message
	color.Set(color.FgHiRed)
	fmt.Fprintf(os.Stderr, "[ERROR]   ")

	color.Set(color.FgHiCyan)
	fmt.Fprintf(os.Stderr, "%s %d:%d ", *file, line, startChar)

	color.Set(color.FgHiWhite)
	fmt.Fprintf(os.Stderr, "%s\n\n", message)
}

func ErrorCodePos(codePos *data.CodePos, message string) {
	ErrorPos(codePos.File, codePos.StartLine, codePos.StartChar, codePos.EndChar, message)
}

func Error2CodePos(codePos1, codePos2 *data.CodePos, message string) {
	// Print error line
	lineString, err := readLine(*codePos1.File, codePos1.StartLine)

	if err == nil {
		color.Set(color.FgWhite)

		if codePos1.StartChar == codePos1.EndChar {
			codePos1.EndChar++
		}
		if codePos2.StartChar == codePos2.EndChar {
			codePos2.EndChar++
		}

		// Print line of code with errors
		fmt.Fprint(os.Stderr, "\t\t ")
		fmt.Fprintf(os.Stderr, "%s", lineString)
		fmt.Fprint(os.Stderr, "\n\t\t")

		// Move to error token 1
		var i uint
		for i = 0; i < codePos1.StartChar; i++ {
			fmt.Fprintf(os.Stderr, " ")
		}

		// Draw arrows under the error token 1
		color.Set(color.FgHiRed)
		for i = codePos1.StartChar; i < codePos1.EndChar; i++ {
			fmt.Fprintf(os.Stderr, "^")
		}

		// Move to error token 2
		for i = codePos1.EndChar; i < codePos2.StartChar; i++ {
			fmt.Fprintf(os.Stderr, " ")
		}

		// Draw arrows under the error token 2
		for i = codePos2.StartChar; i < codePos2.EndChar; i++ {
			fmt.Fprintf(os.Stderr, "^")
		}

		fmt.Fprintf(os.Stderr, "\n")
	}

	// Print message
	fmt.Fprintf(os.Stderr, "[ERROR]   ")

	color.Set(color.FgHiCyan)
	fmt.Fprintf(os.Stderr, "%s %d:%d ", *codePos1.File, codePos1.StartLine, codePos1.StartChar)

	color.Set(color.FgHiWhite)
	fmt.Fprintf(os.Stderr, "%s\n\n", message)
}

func Fatal(error_code int, message string) {
	color.Set(color.FgHiRed)
	color.Set(color.Bold)
	fmt.Fprintf(os.Stderr, "[FATAL]   %s\n", message)
	color.Set(color.Reset)

	os.Exit(error_code)
}
