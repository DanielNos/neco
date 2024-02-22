package logger

import (
	"bufio"
	"fmt"
	data "neco/dataStructures"
	"os"

	color "github.com/fatih/color"
)

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
	color.Set(color.FgHiGreen)
	fmt.Print("[SUCCESS] ")
	color.Set(color.FgHiWhite)

	fmt.Println(message)
}

func Info(message string) {
	color.Set(color.FgHiWhite)
	fmt.Print("[INFO]    ")

	fmt.Println(message)
}

func Warning(message string) {
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
	// Print error line
	lineString, err := readLine(*file, line)

	if err == nil {
		color.Set(color.FgWhite)

		if startChar == endChar {
			endChar++
		}

		// Print line of code with error
		fmt.Fprint(os.Stderr, "\t\t")
		fmt.Fprintf(os.Stderr, "%s", lineString)
		fmt.Fprint(os.Stderr, "\n\t\t")

		// Move to error token
		var i uint
		for i = 0; i < startChar-1; i++ {
			if lineString[i] == '\t' {
				fmt.Fprintf(os.Stderr, "\t")
			} else {
				fmt.Fprintf(os.Stderr, " ")
			}
		}

		// Draw arrows under the error token
		color.Set(color.FgHiRed)
		for i = startChar; i < endChar; i++ {
			fmt.Fprintf(os.Stderr, "^")
		}

		fmt.Fprintf(os.Stderr, "\n")
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
	ErrorPos(codePos.File, codePos.Line, codePos.StartChar, codePos.EndChar, message)
}

func Error2CodePos(codePos1, codePos2 *data.CodePos, message string) {
	// Print error line
	lineString, err := readLine(*codePos1.File, codePos1.Line)

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
	fmt.Fprintf(os.Stderr, "%s %d:%d ", *codePos1.File, codePos1.Line, codePos1.StartChar)

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
