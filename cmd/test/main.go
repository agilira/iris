package main

import (
	"fmt"
	"os"
	"time"

	"github.com/agilira/iris"
)

func main() {
	logger, err := iris.New(iris.Config{
		Level:   iris.Debug,
		Encoder: iris.NewJSONEncoder(),
		Output:  os.Stdout,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			fmt.Printf("Error closing logger: %v\n", err)
		}
	}()

	fmt.Println("Logger created")
	logger.Start()
	fmt.Println("Logger started")

	result := logger.Info("test message", iris.Str("key", "value"))
	fmt.Printf("Log result: %v\n", result)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	if err := logger.Sync(); err != nil {
		fmt.Printf("Warning: failed to sync logger: %v\n", err)
	}
	fmt.Println("Logger synced")
}
