package logger

import (
	"bufio"
	"fmt"
	"neko/dataStructures"
	"os"

	color "github.com/fatih/color"
)

func readLine(filePath string, lineIndex uint) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentLine uint = 0

	for scanner.Scan() {
		currentLine++
		if currentLine == lineIndex {
			return scanner.Text(), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %s", err)
	}

	return "", fmt.Errorf("line number out of range")
}

func Success(message string) {
	color.Set(color.FgHiGreen)
	fmt.Print( "[SUCCESS] ");
	color.Set(color.FgHiWhite)
	
	fmt.Println(message)
}

func Info(message string) {
	color.Set(color.FgHiWhite)
	fmt.Print( "[INFO] ");
	
	fmt.Println(message)
}

func Warning(message string) {
	color.Set(color.FgHiYellow)
	fmt.Print( "[WARNING] ");
	color.Set(color.FgHiWhite)

	fmt.Println(message)
}

func Error(message string) {
	color.Set(color.FgHiRed)
	fmt.Fprint(os.Stderr, "[ERROR] ");
	color.Set(color.FgHiWhite)

	fmt.Fprintln(os.Stderr, message)
}

func ErrorPos(file *string, line, startChar, endChar uint, message string) {
	// Print error line
	lineString, err := readLine(*file, line)

	if err == nil {
		color.Set(color.FgWhite)
		
		// Print line of code with error
		fmt.Fprint(os.Stderr, "\t\t ")
		fmt.Fprintf(os.Stderr, "%s", lineString)
		fmt.Fprint(os.Stderr, "\n\t\t")

		// Move to error token
		var i uint
		for i = 0; i < startChar; i++ {
			fmt.Fprintf(os.Stderr, " ");
		}

		// Draw arrows under the error token
		color.Set(color.FgHiRed)
		for i = startChar; i < endChar; i++ {
			fmt.Fprintf(os.Stderr, "^");
		}

		fmt.Fprintf(os.Stderr, "\n");
	}

	// Print message
	fmt.Fprintf(os.Stderr, "[ERROR] ")
	
	color.Set(color.FgHiCyan)
	fmt.Fprintf(os.Stderr, "%s %d:%d ", *file, line, startChar)

	color.Set(color.FgHiWhite)
	fmt.Fprintf(os.Stderr, "%s\n\n", message)
}

func ErrorCodePos(codePos *dataStructures.CodePos, message string) {
	ErrorPos(codePos.File, codePos.Line, codePos.StartChar, codePos.EndChar, message)
}

func Fatal(error_code int, message string) {
	Error(message)
	os.Exit(error_code)
}
