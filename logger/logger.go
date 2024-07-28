package logger

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	data "github.com/DanielNos/neco/dataStructures"

	color "github.com/fatih/color"
)

const (
	LL_Info byte = iota
	LL_Success
	LL_Warning
	LL_Error
	LL_Fatal
	LL_NoLog
)

var StringToLogLevel = map[string]byte{
	"info":    LL_Info,
	"success": LL_Success,
	"warning": LL_Warning,
	"error":   LL_Warning,
	"fatal":   LL_Fatal,
	"nolog":   LL_NoLog,
}

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
	if LoggingLevel > LL_Success {
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

func WarningPos(file *string, line, startChar, endChar uint, message string) {
	if LoggingLevel > LL_Warning {
		return
	}

	// Read line
	lineString, err := readLine(*file, line)

	if int(endChar) > len(lineString) {
		endChar--
	}

	// Print line
	if err == nil && len(strings.Trim(lineString, "\n \t")) != 0 {
		// Print from start to error
		color.Set(color.FgWhite)

		fmt.Fprint(os.Stderr, lineString[0:startChar-1])

		// Print error
		color.Set(color.FgHiRed)
		color.Set(color.Underline)

		fmt.Fprint(os.Stderr, lineString[startChar-1:endChar])

		// Print from error to end
		color.Set(color.Reset)
		color.Set(color.FgWhite)

		fmt.Fprint(os.Stderr, lineString[endChar:])
		fmt.Fprintln(os.Stderr, "\n")
	}

	// Print message
	color.Set(color.FgHiYellow)
	fmt.Print("[WARNING]   ")

	color.Set(color.FgHiCyan)
	fmt.Fprintf(os.Stderr, "%s %d:%d ", *file, line, startChar)

	color.Set(color.FgHiWhite)
	fmt.Fprint(os.Stderr, message+"\n\n")
}

func WarningCodePos(codePos *data.CodePos, message string) {
	WarningPos(codePos.File, codePos.StartLine, codePos.StartChar, codePos.EndChar, message)
}

func Error(message string) {
	if LoggingLevel > LL_Error {
		return
	}

	color.Set(color.FgHiRed)
	fmt.Fprint(os.Stderr, "[ERROR]   ")
	color.Set(color.FgHiWhite)

	fmt.Fprintln(os.Stderr, message)
}

func ErrorPos(file *string, line, startChar, endChar uint, message string) {
	if LoggingLevel > LL_Error {
		return
	}

	// Read line
	lineString, err := readLine(*file, line)

	if int(endChar) > len(lineString) {
		endChar--
	}

	// Print line
	if err == nil && len(strings.Trim(lineString, "\n \t")) != 0 {
		// Print from start to error
		color.Set(color.FgWhite)

		fmt.Fprint(os.Stderr, lineString[0:startChar-1])

		// Print error
		color.Set(color.FgHiRed)
		color.Set(color.Underline)

		fmt.Fprint(os.Stderr, lineString[startChar-1:endChar])

		// Print from error to end
		color.Set(color.Reset)
		color.Set(color.FgWhite)

		fmt.Fprint(os.Stderr, lineString[endChar:])
		fmt.Fprintln(os.Stderr, "\n")
	}

	// Print message
	color.Set(color.FgHiRed)
	fmt.Fprint(os.Stderr, "[ERROR]   ")

	color.Set(color.FgHiCyan)
	fmt.Fprintf(os.Stderr, "%s %d:%d ", *file, line, startChar)

	color.Set(color.FgHiWhite)
	fmt.Fprint(os.Stderr, message+"\n\n")
}

func ErrorCodePos(codePos *data.CodePos, message string) {
	ErrorPos(codePos.File, codePos.StartLine, codePos.StartChar, codePos.EndChar, message)
}

func Error2CodePos(codePos1, codePos2 *data.CodePos, message string) {
	if LoggingLevel > LL_Error {
		return
	}

	// Read line
	lineString, err := readLine(*codePos1.File, codePos1.StartLine)

	if int(codePos1.EndChar) > len(lineString) {
		codePos1.EndChar--
	}

	if int(codePos2.EndChar) > len(lineString) {
		codePos2.EndChar--
	}

	// Print line
	if err == nil && len(strings.Trim(lineString, "\n \t")) != 0 {
		// Print from start to error 1
		color.Set(color.FgWhite)

		fmt.Fprint(os.Stderr, lineString[0:codePos1.StartChar-1])

		// Print error 1
		color.Set(color.FgHiRed)
		color.Set(color.Underline)

		fmt.Fprint(os.Stderr, lineString[codePos1.StartChar-1:codePos1.EndChar])

		// Print from end of error 1 to start of error 2
		color.Set(color.Reset)
		color.Set(color.FgWhite)

		fmt.Fprint(os.Stderr, lineString[codePos1.EndChar:codePos2.StartChar-1])

		// Print error 2
		color.Set(color.FgHiRed)
		color.Set(color.Underline)

		fmt.Fprint(os.Stderr, lineString[codePos2.StartChar-1:codePos2.EndChar])

		// Print from error 2 to end
		color.Set(color.Reset)
		color.Set(color.FgWhite)

		fmt.Fprint(os.Stderr, lineString[codePos2.EndChar:])
		fmt.Fprintln(os.Stderr, "\n")
	}

	// Print message
	color.Set(color.FgHiRed)
	fmt.Fprint(os.Stderr, "[ERROR]   ")

	color.Set(color.FgHiCyan)
	fmt.Fprintf(os.Stderr, "%s %d:%d ", *codePos1.File, codePos1.StartLine, codePos1.StartChar)

	color.Set(color.FgHiWhite)
	fmt.Fprint(os.Stderr, message+"\n\n")
}

func Fatal(error_code int, message string) {
	if LoggingLevel > LL_Fatal {
		os.Exit(error_code)
	}

	color.Set(color.FgHiRed)
	color.Set(color.Bold)
	fmt.Fprint(os.Stderr, "[FATAL]   "+message+"\n")
	color.Set(color.Reset)

	os.Exit(error_code)
}
