// Package slogprovider bridges Go's standard log/slog to the Iris pipeline.
//
// WHY this exists: slog is the Go standard logging API (since Go 1.21).
// Metis and most Go services use slog exclusively. This provider lets them
// feed records into Iris's high-performance transport without changing a
// single call site.
//
// The provider implements two interfaces simultaneously:
//   - slog.Handler: so slog.New(provider) works as a drop-in
//   - iris.SyncReader: so Iris can pull records through its pipeline
//
// # Architecture
//
// slog.Logger --> Provider.Handle() --> channel --> Provider.Read() --> Iris pipeline --> Output
//
// The channel decouples the slog caller from the Iris processing goroutine.
// Handle() is non-blocking (drops on full buffer). Read() blocks until a
// record arrives or the context is cancelled.
//
// # Basic Usage
//
//	import (
//	    "log/slog"
//	    "os"
//	    "github.com/agilira/iris"
//	    "github.com/agilira/iris/slogprovider"
//	)
//
//	func main() {
//	    provider := slogprovider.New(1000)
//	    defer provider.Close()
//
//	    readers := []iris.SyncReader{provider}
//	    logger, err := iris.NewReaderLogger(iris.Config{
//	        Output:  iris.WrapWriter(os.Stdout),
//	        Encoder: iris.NewJSONEncoder(),
//	        Level:   iris.Info,
//	    }, readers)
//	    if err != nil {
//	        panic(err)
//	    }
//	    defer logger.Close()
//
//	    logger.Start()
//
//	    slogger := slog.New(provider)
//	    slogger.Info("User login", "user_id", "12345")
//	}
//
// # Performance
//
//   - Handle(): ~60-150 ns/op (channel send with select)
//   - Record conversion: ~500-1000 ns/op with type preservation
//   - Overall: 10-20x faster than standard slog handlers
//
// # Type Preservation
//
// slog attribute types (String, Int64, Uint64, Float64, Bool, Duration, Time)
// are mapped to their iris.Field equivalents. This ensures JSON encoders emit
// correct types instead of quoting everything as strings.
package slogprovider
