// batch_processor.go: High-performance batch log conversion with worker pool
//
// Copyright (c) 2025 AGILira
// Series: an AGILira fragment
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// FileTask represents a file conversion task
type FileTask struct {
	InputPath  string
	OutputPath string
	Config     *Config
}

// BatchProcessor uses worker pool for ultra-fast parallel log conversion
type BatchProcessor struct {
	converter *BinaryToJSONConverter
	config    *Config
	stats     *BatchStats
	mu        sync.RWMutex
}

// BatchStats tracks conversion statistics
type BatchStats struct {
	FilesProcessed int64
	FilesError     int64
	BytesProcessed int64
	StartTime      time.Time
	EndTime        time.Time
}

// NewBatchProcessor creates a new high-performance batch processor
func NewBatchProcessor(config *Config) (*BatchProcessor, error) {
	return &BatchProcessor{
		converter: NewBinaryToJSONConverterWithOptions(config.Pretty, config.LevelFilter, config.ValidateOnly),
		config:    config,
		stats: &BatchStats{
			StartTime: time.Now(),
		},
	}, nil
}

// ProcessDirectory processes an entire directory using worker pool parallel processing
func (bp *BatchProcessor) ProcessDirectory(inputDir, outputDir string) error {
	// Calculate optimal worker count and setup
	workers := runtime.NumCPU()
	if bp.config.Verbose {
		fmt.Fprintf(os.Stderr, "Initializing batch processor with %d workers\n", workers)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0750); err != nil { // Secure directory permissions
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	return bp.processWithWorkerPool(inputDir, outputDir, workers)
}

func (bp *BatchProcessor) processWithWorkerPool(inputDir, outputDir string, workers int) error {
	// Create task channel with buffer
	taskChan := make(chan *FileTask, workers*2)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go bp.worker(taskChan, &wg)
	}

	// Scan directory and queue tasks
	taskCount := bp.queueTasks(inputDir, outputDir, taskChan)

	if bp.config.Verbose {
		fmt.Fprintf(os.Stderr, "Processing files with %d workers...\n", workers)
	}

	// Wait for all workers to complete
	wg.Wait()

	if bp.config.Verbose {
		fmt.Fprintf(os.Stderr, "Completed processing %d files\n", taskCount)
	}
	return nil
}

// worker processes file conversion tasks
func (bp *BatchProcessor) worker(taskChan <-chan *FileTask, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskChan {
		if err := bp.convertSingleFile(task); err != nil {
			bp.mu.Lock()
			bp.stats.FilesError++
			bp.mu.Unlock()

			if bp.config.Verbose {
				fmt.Fprintf(os.Stderr, "Error converting %s: %v\n", task.InputPath, err)
			}
			continue
		}

		bp.mu.Lock()
		bp.stats.FilesProcessed++
		bp.mu.Unlock()
	}
}

// convertSingleFile converts a single file
func (bp *BatchProcessor) convertSingleFile(task *FileTask) error {
	// Open input file
	input, err := os.Open(task.InputPath)
	if err != nil {
		return fmt.Errorf("failed to open input: %v", err)
	}
	defer input.Close()

	// Create output directory if needed
	if dir := filepath.Dir(task.OutputPath); dir != "." {
		if err := os.MkdirAll(dir, 0750); err != nil { // Secure directory permissions
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	// Create output file
	output, err := os.Create(task.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output: %v", err)
	}
	defer output.Close()

	// Get file size for stats
	if info, err := input.Stat(); err == nil {
		bp.mu.Lock()
		bp.stats.BytesProcessed += info.Size()
		bp.mu.Unlock()
	}

	// Convert using existing converter
	return bp.converter.Convert(input, output)
}

// isLogFile determines if a file should be processed
func (bp *BatchProcessor) isLogFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".log" || ext == ".iris" || ext == ".txt"
}

func (bp *BatchProcessor) queueTasks(inputDir, outputDir string, taskChan chan<- *FileTask) int {
	taskCount := 0
	go func() {
		defer close(taskChan)

		err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			if bp.processFile(inputDir, outputDir, path, taskChan, &taskCount) != nil {
				return err
			}

			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		}
	}()
	return taskCount
}

func (bp *BatchProcessor) processFile(inputDir, outputDir, path string, taskChan chan<- *FileTask, taskCount *int) error {
	// Filter log files
	if !bp.isLogFile(path) {
		if bp.config.Verbose {
			fmt.Fprintf(os.Stderr, "Skipping %s (not a log file)\n", path)
		}
		return nil
	}

	// Calculate output path
	task, err := bp.createFileTask(inputDir, outputDir, path)
	if err != nil {
		return err
	}

	// Send to worker pool
	taskChan <- task
	*taskCount++

	if bp.config.Verbose && *taskCount%100 == 0 {
		fmt.Fprintf(os.Stderr, "Queued %d tasks...\n", *taskCount)
	}

	return nil
}

func (bp *BatchProcessor) createFileTask(inputDir, outputDir, path string) (*FileTask, error) {
	// Calculate output path
	relPath, err := filepath.Rel(inputDir, path)
	if err != nil {
		return nil, err
	}
	outputPath := filepath.Join(outputDir, strings.TrimSuffix(relPath, filepath.Ext(relPath))+".json")

	// Create file task
	return &FileTask{
		InputPath:  path,
		OutputPath: outputPath,
		Config:     bp.config,
	}, nil
}
