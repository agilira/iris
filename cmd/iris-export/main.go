// iris-export: CLI tool for converting Iris binary logs to JSON
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	version = "1.0.0"
	usage   = `iris-export - Convert Iris binary logs to JSON format

USAGE:
    iris-export [OPTIONS]

EXAMPLES:
    # Convert single file
    iris-export -i app.log -o app.json
    
    # Stream from stdin to stdout
    ./myapp | iris-export > app.json
    iris-export < app.log > app.json
    
    # Batch convert directory
    iris-export -i logs/ -o json/ -r
    
    # Filter by log level
    iris-export -i app.log -level error -o errors.json
    
    # Validate format only
    iris-export -i app.log -validate

OPTIONS:
`
)

type Config struct {
	Input        string
	Output       string
	Recursive    bool
	Pretty       bool
	Verbose      bool
	Version      bool
	LevelFilter  string // Filter by log level (debug, info, warn, error, etc.)
	ValidateOnly bool   // Only validate input format without converting
}

func main() {
	config := parseFlags()

	if config.Version {
		fmt.Printf("iris-export version %s\n", version)
		os.Exit(0)
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.Input, "i", "", "Input file or directory (use '-' or empty for stdin)")
	flag.StringVar(&config.Input, "input", "", "Input file or directory (use '-' or empty for stdin)")
	flag.StringVar(&config.Output, "o", "", "Output file or directory (use '-' or empty for stdout)")
	flag.StringVar(&config.Output, "output", "", "Output file or directory (use '-' or empty for stdout)")
	flag.BoolVar(&config.Recursive, "r", false, "Recursively process directories")
	flag.BoolVar(&config.Recursive, "recursive", false, "Recursively process directories")
	flag.BoolVar(&config.Pretty, "p", false, "Pretty-print JSON output")
	flag.BoolVar(&config.Pretty, "pretty", false, "Pretty-print JSON output")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&config.Version, "version", false, "Show version information")
	flag.StringVar(&config.LevelFilter, "level", "", "Filter by log level (debug, info, warn, error, dpanic, panic, fatal)")
	flag.BoolVar(&config.ValidateOnly, "validate", false, "Only validate input format without converting")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSupported log levels: %v\n", GetAllLevelNames())
	}

	flag.Parse()

	// Validate level filter if provided
	if config.LevelFilter != "" {
		validLevels := GetAllLevelNames()
		valid := false
		for _, level := range validLevels {
			if config.LevelFilter == level {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Fprintf(os.Stderr, "Error: Invalid level '%s'. Valid levels: %v\n", config.LevelFilter, validLevels)
			os.Exit(1)
		}
	}

	return config
}

func run(config *Config) error {
	// Handle stdin/stdout case
	if config.Input == "" || config.Input == "-" {
		return handleStdinInput(config)
	}

	// Check if input exists
	info, err := os.Stat(config.Input)
	if err != nil {
		return fmt.Errorf("input path not found: %v", err)
	}

	// Handle directory case
	if info.IsDir() {
		return handleDirectoryInput(config)
	}

	// Handle single file case
	return handleFileInput(config)
}

func handleStdinInput(config *Config) error {
	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Reading from stdin...\n")
	}
	return convertStream(os.Stdin, os.Stdout, config)
}

func handleDirectoryInput(config *Config) error {
	if config.Output == "" || config.Output == "-" {
		return fmt.Errorf("directory input requires output directory")
	}

	// Use Zetes batch processor for directory conversion
	batchProcessor, err := NewBatchProcessor(config)
	if err != nil {
		return fmt.Errorf("failed to create batch processor: %v", err)
	}

	return batchProcessor.ProcessDirectory(config.Input, config.Output)
}

func handleFileInput(config *Config) error {
	if config.Output == "" || config.Output == "-" {
		// Single file to stdout
		return handleFileToStdout(config)
	}

	// Single file to file
	return convertFile(config.Input, config.Output, config)
}

func handleFileToStdout(config *Config) error {
	input, err := os.Open(config.Input)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer input.Close()

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Converting %s to stdout...\n", config.Input)
	}
	return convertStream(input, os.Stdout, config)
}

func convertStream(input io.Reader, output io.Writer, config *Config) error {
	converter := NewBinaryToJSONConverterWithOptions(config.Pretty, config.LevelFilter, config.ValidateOnly)
	return converter.Convert(input, output)
}

func convertFile(inputPath, outputPath string, config *Config) error {
	// #nosec G304 - File paths are provided via CLI arguments, validated by application
	input, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer input.Close()

	// Create output directory if needed
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0750); err != nil { // Secure directory permissions
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	// #nosec G304 - Output path is provided via CLI arguments, validated by application
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer output.Close()

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Converting %s -> %s\n", inputPath, outputPath)
	}

	return convertStream(input, output, config)
}
