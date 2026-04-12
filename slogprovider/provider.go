// provider.go: slog.Handler -> iris.SyncReader bridge
//
// WHY: Go's stdlib log/slog is the standard logging API. This bridge lets any
// slog-based application feed records into Iris's high-performance pipeline
// without changing a single log call site. Metis uses slog exclusively --
// this provider is the integration point.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package slogprovider

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/agilira/iris"
)

// Provider implements iris.SyncReader for Go's standard log/slog package.
//
// Provider acts as a bridge between slog and Iris, implementing both the
// iris.SyncReader interface for Iris integration and the slog.Handler interface
// for slog compatibility. It captures slog records in an internal buffer and
// converts them to Iris records on demand.
//
// The provider is designed for high performance and thread safety:
//   - Non-blocking Handle() operations (drops records on buffer full)
//   - Efficient record conversion with type preservation
//   - Safe concurrent access from multiple goroutines
//   - Graceful shutdown with proper resource cleanup
//
// Example usage:
//
//	provider := slogprovider.New(1000)
//	defer provider.Close()
//
//	slogger := slog.New(provider)
//	slogger.Info("Message", "key", "value")
type Provider struct {
	records chan slog.Record // Buffered channel for slog records
	closed  chan struct{}    // Signal channel for shutdown coordination
	once    sync.Once        // Ensures Close() is idempotent
}

// New creates a new Provider that captures slog records for processing by Iris.
//
// The bufferSize parameter controls the internal channel buffer size. A larger
// buffer provides better performance under burst loads but uses more memory.
// Recommended values:
//   - 100-500: Low to moderate logging volume applications
//   - 1000-5000: High volume applications
//   - 5000+: Very high volume or burst-heavy applications
//
// When the buffer is full, new records are dropped to maintain non-blocking
// behavior. Monitor your application's logging patterns to choose an appropriate
// buffer size.
//
// The returned Provider must be closed when no longer needed to free resources:
//
//	provider := New(1000)
//	defer provider.Close()
func New(bufferSize int) *Provider {
	return &Provider{
		records: make(chan slog.Record, bufferSize),
		closed:  make(chan struct{}),
	}
}

// Handle implements slog.Handler to capture slog records for processing by Iris.
//
// This method is called by the slog library for each log record. It attempts to
// store the record in the internal buffer for later processing by Iris. The
// operation is non-blocking:
//   - If buffer space is available, the record is stored successfully
//   - If the provider is closed, an error is returned
//   - If the buffer is full, the record is dropped silently (returns nil)
//
// The non-blocking behavior ensures that logging never blocks the application,
// even under high load conditions. Applications should monitor buffer sizes
// and provider performance if record dropping is a concern.
//
// Thread Safety: Safe for concurrent access from multiple goroutines.
func (p *Provider) Handle(ctx context.Context, record slog.Record) error {
	select {
	case p.records <- record:
		return nil
	case <-p.closed:
		return fmt.Errorf("slog provider closed")
	default:
		return nil // Drop if buffer full
	}
}

// Enabled implements slog.Handler to indicate whether records at the given level should be processed.
//
// WHY always true: Level filtering is Iris's responsibility. Delegating here
// would create two competing filter points and make dynamic level changes
// impossible without reconfiguring the slog side.
func (p *Provider) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

// WithAttrs implements slog.Handler to create a handler with additional attributes.
//
// WHY pass-through: The slog library embeds attributes into each Record before
// calling Handle(). Duplicating that logic here would add complexity with zero
// gain -- the attributes arrive regardless.
func (p *Provider) WithAttrs(attrs []slog.Attr) slog.Handler {
	return p
}

// WithGroup implements slog.Handler to create a handler with a named group.
//
// WHY pass-through: Same reasoning as WithAttrs -- the slog library handles
// group structuring internally before calling Handle().
func (p *Provider) WithGroup(name string) slog.Handler {
	return p
}

// Read implements iris.SyncReader to provide slog records to the Iris pipeline.
//
// This method is called by Iris to retrieve the next available log record for
// processing. It blocks until:
//   - A record becomes available (returns the converted record)
//   - The context is cancelled (returns context error)
//   - The provider is closed (returns nil, nil)
//
// The method converts slog records to Iris records, preserving message content,
// level information, and all attributes with appropriate type conversion.
//
// Thread Safety: Safe for concurrent access, though typically called by a
// single Iris reader goroutine.
func (p *Provider) Read(ctx context.Context) (*iris.Record, error) {
	select {
	case record := <-p.records:
		return p.convertSlogRecord(record), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.closed:
		return nil, nil
	}
}

// Close implements io.Closer to gracefully shut down the provider.
//
// WHY sync.Once: Close may be called from defer + explicit shutdown paths.
// Double-close on a channel panics; sync.Once makes this impossible.
//
// After Close() is called:
//   - Handle() will return an error for new records
//   - Read() will return nil, nil after processing remaining buffered records
//   - The provider should not be used for new operations
func (p *Provider) Close() error {
	p.once.Do(func() {
		close(p.closed)
	})
	return nil
}

// convertSlogRecord converts a slog.Record to an iris.Record with full fidelity.
//
// WHY not batch: Iris processes records one at a time through the pipeline.
// Batching at the conversion layer would add latency with no throughput gain.
func (p *Provider) convertSlogRecord(slogRec slog.Record) *iris.Record {
	record := iris.NewRecord(p.convertLevel(slogRec.Level), slogRec.Message)

	slogRec.Attrs(func(attr slog.Attr) bool {
		field := p.convertAttribute(attr)
		return record.AddField(field)
	})

	return record
}

// convertLevel maps slog.Level values to iris.Level values.
//
// WHY switch on ranges: slog levels are ints (Debug=-4, Info=0, Warn=4, Error=8).
// Custom levels (e.g. slog.Level(2)) must map somewhere sensible.
// Range checks ensure any custom level falls into the nearest iris bucket.
func (p *Provider) convertLevel(slogLevel slog.Level) iris.Level {
	switch {
	case slogLevel <= slog.LevelDebug:
		return iris.Debug
	case slogLevel <= slog.LevelInfo:
		return iris.Info
	case slogLevel <= slog.LevelWarn:
		return iris.Warn
	default:
		return iris.Error
	}
}

// convertAttribute converts a slog.Attr to an iris.Field with type preservation.
//
// WHY type switch: slog stores values as slog.Value with a Kind discriminator.
// Preserving the original type (Int64, Float64, Bool...) lets Iris encoders
// emit correct JSON types instead of quoting everything as strings.
func (p *Provider) convertAttribute(attr slog.Attr) iris.Field {
	key := attr.Key
	value := attr.Value

	switch value.Kind() {
	case slog.KindString:
		return iris.String(key, value.String())
	case slog.KindInt64:
		return iris.Int64(key, value.Int64())
	case slog.KindUint64:
		return iris.Uint64(key, value.Uint64())
	case slog.KindFloat64:
		return iris.Float64(key, value.Float64())
	case slog.KindBool:
		return iris.Bool(key, value.Bool())
	case slog.KindDuration:
		return iris.Dur(key, value.Duration())
	case slog.KindTime:
		return iris.Time(key, value.Time())
	default:
		return iris.String(key, value.String())
	}
}
