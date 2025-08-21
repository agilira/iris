// writer.go: Output writers for Iris
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package iris

import (
	"io"
	"os"
)

// Writer interface for log output destinations
type Writer interface {
	Write([]byte) (int, error)
}

// ConsoleWriter wraps an io.Writer for console output
type ConsoleWriter struct {
	writer io.Writer
}

// NewConsoleWriter creates a new console writer
func NewConsoleWriter(w io.Writer) *ConsoleWriter {
	return &ConsoleWriter{writer: w}
}

// Write implements the Writer interface
func (c *ConsoleWriter) Write(data []byte) (int, error) {
	return c.writer.Write(data)
}

// Common writer instances
var (
	// StdoutWriter writes to standard output
	StdoutWriter = NewConsoleWriter(os.Stdout)
	// StderrWriter writes to standard error
	StderrWriter = NewConsoleWriter(os.Stderr)
)

// FileWriter writes to a file
type FileWriter struct {
	file *os.File
	path string
}

// NewFileWriter creates a new file writer
func NewFileWriter(path string) (*FileWriter, error) {
	// #nosec G304 - File path is provided by application, not user input
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600) // Secure permissions
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		file: file,
		path: path,
	}, nil
}

// Write implements the Writer interface
func (f *FileWriter) Write(data []byte) (int, error) {
	return f.file.Write(data)
}

// Close closes the file
func (f *FileWriter) Close() error {
	return f.file.Close()
}

// Sync syncs the file to disk
func (f *FileWriter) Sync() error {
	return f.file.Sync()
}
