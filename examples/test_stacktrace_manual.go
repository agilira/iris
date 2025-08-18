// test_stacktrace_manual.go - Manual test for stack trace presets
// go run test_stacktrace_manual.go

package main

import (
	"fmt"
	"log"

	"github.com/agilira/iris"
)

func main() {
	fmt.Println("Testing Development with Stack Trace preset:")

	logger, err := iris.NewDevelopmentWithStackTrace()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close()

	logger.Info("This is an info message - no stack trace expected")
	logger.Warn("This is a warning message - no stack trace expected")
	logger.Error("This is an error message - stack trace expected")

	fmt.Println("\nTesting Debug with Stack Trace preset:")

	debugLogger, err := iris.NewDebugWithStackTrace()
	if err != nil {
		log.Fatal(err)
	}
	defer debugLogger.Close()

	debugLogger.Info("This is an info message - no stack trace expected")
	debugLogger.Warn("This is a warning message - stack trace expected")
	debugLogger.Error("This is an error message - stack trace expected")
}
