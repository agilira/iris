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

OPTIONS:
`
)

type Config struct {
	Input     string
	Output    string
	Recursive bool
	Pretty    bool
	Verbose   bool
	Version   bool
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

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}

	flag.Parse()

	return config
}

func run(config *Config) error {
	// Handle stdin/stdout case
	if config.Input == "" || config.Input == "-" {
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "Reading from stdin...\n")
		}
		return convertStream(os.Stdin, os.Stdout, config)
	}

	// Check if input exists
	info, err := os.Stat(config.Input)
	if err != nil {
		return fmt.Errorf("input path not found: %v", err)
	}

	// Handle directory case
	if info.IsDir() {
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

	// Handle single file case
	if config.Output == "" || config.Output == "-" {
		// Single file to stdout
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

	// Single file to file
	return convertFile(config.Input, config.Output, config)
}

func convertStream(input io.Reader, output io.Writer, config *Config) error {
	converter := NewBinaryToJSONConverter(config.Pretty)
	return converter.Convert(input, output)
}

func convertFile(inputPath, outputPath string, config *Config) error {
	input, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer input.Close()

	// Create output directory if needed
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

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
