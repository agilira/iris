// adaptive.go: slog.Handler -> AdaptiveLogger direct bridge
//
// WHY: The existing Provider uses a channel-based SyncReader pattern,
// adding a goroutine hop and buffer between slog and iris. AdaptiveHandler
// eliminates that indirection: Handle() converts the slog.Record and writes
// directly into AdaptiveLogger, which routes through iris's ring buffer.
// This is the production path for Metis's pipeline.
//
// Pipeline:
//   slog.Info() -> AdaptiveHandler.Handle() -> AdaptiveLogger.Info/Debug/Warn/Error()
//                -> iris ring -> processor -> output (lethe / stdout)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package slogprovider

import (
	"context"
	"log/slog"

	"github.com/agilira/iris"
)

// AdaptiveHandler implements slog.Handler by writing directly into an
// AdaptiveLogger. No channels, no goroutine hops -- the slog.Record is
// converted to iris fields and dispatched synchronously into the ring buffer.
//
// Thread Safety: Safe for concurrent use. AdaptiveLogger handles all
// synchronization internally via atomic contention detection.
type AdaptiveHandler struct {
	al *iris.AdaptiveLogger
}

// NewAdaptiveHandler creates a handler that bridges slog to an AdaptiveLogger.
// The caller owns the AdaptiveLogger lifecycle (Start, Close, Sync).
func NewAdaptiveHandler(al *iris.AdaptiveLogger) *AdaptiveHandler {
	return &AdaptiveHandler{al: al}
}

// Handle implements slog.Handler. Converts the slog.Record into iris fields
// and dispatches to the appropriate AdaptiveLogger method based on level.
//
// WHY synchronous: the AdaptiveLogger's ring buffer is the async boundary.
// Adding another buffer here would defeat the purpose (latency + complexity).
func (h *AdaptiveHandler) Handle(_ context.Context, record slog.Record) error {
	fields := make([]iris.Field, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, convertAttribute(attr))
		return true
	})

	// WHY switch on ranges (not == ): slog levels are ints; custom levels
	// (e.g. slog.Level(2)) must map to the nearest iris bucket.
	switch {
	case record.Level <= slog.LevelDebug:
		h.al.Debug(record.Message, fields...)
	case record.Level <= slog.LevelInfo:
		h.al.Info(record.Message, fields...)
	case record.Level <= slog.LevelWarn:
		h.al.Warn(record.Message, fields...)
	default:
		h.al.Error(record.Message, fields...)
	}
	return nil
}

// Enabled implements slog.Handler. Always returns true because level filtering
// is iris's responsibility. Returning false here would prevent records from
// reaching iris, making dynamic level changes via SetLevel impossible.
func (h *AdaptiveHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// WithAttrs implements slog.Handler. Returns self because slog embeds
// attributes into each Record before calling Handle(). Duplicating that
// logic here adds complexity with zero gain.
func (h *AdaptiveHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup implements slog.Handler. Returns self -- same reasoning as WithAttrs.
func (h *AdaptiveHandler) WithGroup(_ string) slog.Handler {
	return h
}
