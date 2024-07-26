package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/DanielNos/NeCo/errors"
	"github.com/DanielNos/NeCo/logger"
)

type Action byte

const (
	A_Build Action = iota
	A_Run
	A_Analyze
	A_BuildAndRun
)

type Configuration struct {
	PrintTokens       bool
	DrawTree          bool
	PrintInstructions bool
	Optimize          bool
	Silent            bool
	PrintConstants    bool

	Action     Action
	TargetPath string
	OutputPath string
}

func processArguments() *Configuration {
	args := os.Args[1:]

	// No action
	if len(args) == 0 {
		logger.Fatal(errors.INVALID_FLAGS, "No action specified. Use neco help for more info.")
	}

	argumentsStart := 2

	// Collect target
	configuration := &Configuration{Optimize: true}

	switch args[0] {
	case "build", "run", "analyze":
		if len(args) == 1 {
			logger.Fatal(errors.INVALID_FLAGS, "No target specified.")
		}
		configuration.TargetPath = args[1]

		switch args[0] {
		case "build":
			configuration.Action = A_Build

		case "run":
			configuration.Action = A_Run

		case "analyze":
			configuration.Action = A_Analyze
		}

	case "help", "--help", "-h":
		printHelp()
		os.Exit(0)

	default:
		if strings.HasSuffix(args[0], ".neco") {
			configuration.Action = A_BuildAndRun
			logger.LoggingLevel = logger.LL_Warning

		} else {
			configuration.Action = A_Run
		}

		configuration.TargetPath = args[0]
		argumentsStart = 1
	}

	// Collect flags
	switch configuration.Action {
	// Build flags
	case A_Build, A_BuildAndRun:
		for i := argumentsStart; i < len(args); i++ {
			switch args[i] {
			case "--tokens", "-to":
				configuration.PrintTokens = true

			case "--tree", "-tr":
				configuration.DrawTree = true

			case "--instructions", "-i":
				configuration.PrintInstructions = true

			case "--dont-optimize", "-d":
				configuration.Optimize = false

			case "--silent", "-s":
				logger.LoggingLevel = logger.LL_Error

			case "--no-log", "-n":
				logger.LoggingLevel = logger.LL_NoLog

			case "--log-level", "-l":
				if i+1 == len(args) {
					logger.Fatal(errors.INVALID_FLAGS, "No logging level provided after "+args[i]+" flag.")
				}

				i++
				loggingLevel, err := strconv.Atoi(args[i])

				if err != nil {
					logger.Fatal(errors.INVALID_FLAGS, "Logging level has to be a number.")
				}

				if loggingLevel < 0 || loggingLevel > 5 {
					logger.Fatal(errors.INVALID_FLAGS, "Invalid logging level "+fmt.Sprintf("%d.", loggingLevel))
				}

				logger.LoggingLevel = byte(loggingLevel)

			case "--out", "-o":
				if i+1 == len(args) {
					logger.Fatal(errors.INVALID_FLAGS, "No output path provided after "+args[i]+" flag.")
				}
				i++

				configuration.OutputPath = args[i]

			case "--constants", "-c":
				configuration.PrintConstants = true

			default:
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+args[i]+"\" for action build.")
			}
		}
	// Run flags
	case A_Run:
		for _, flag := range args[argumentsStart:] {
			switch flag {
			default:
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+flag+"\" for action run.")
			}
		}
	// Analyze flags
	case A_Analyze:
		for _, flag := range args[2:] {
			switch flag {
			case "--tokens", "-to":
				configuration.PrintTokens = true

			case "--tree", "-tr":
				configuration.DrawTree = true

			case "--dont-optimize", "-d":
				configuration.Optimize = false

			default:
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+flag+"\" for action analyze.")
			}
		}
	}

	// Set output binary path
	if configuration.OutputPath == "" {
		if strings.HasSuffix(configuration.TargetPath, ".neco") && len(configuration.TargetPath) > 5 {
			configuration.OutputPath = configuration.TargetPath[:len(configuration.TargetPath)-5]

			if runtime.GOOS == "windows" {
				configuration.OutputPath += ".nc"
			}
		} else if runtime.GOOS == "windows" {
			configuration.OutputPath += ".nc"
		} else {
			configuration.OutputPath += "_bin"
		}
	}

	return configuration
}
