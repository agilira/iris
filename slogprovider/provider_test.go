// provider_test.go: Unit tests for slog -> iris bridge
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package slogprovider

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	provider := New(100)
	if provider == nil {
		t.Error("New() returned nil")
	}
	if err := provider.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestProvider_Handle(t *testing.T) {
	provider := New(10)
	defer func() {
		if err := provider.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)

	err := provider.Handle(ctx, record)
	if err != nil {
		t.Errorf("Handle() error = %v", err)
	}
}

func TestProvider_Enabled(t *testing.T) {
	provider := New(100)
	defer func() {
		if err := provider.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	ctx := context.Background()
	if !provider.Enabled(ctx, slog.LevelInfo) {
		t.Error("Expected Enabled to return true for any level")
	}
}

func TestIntegrationWithSlog(t *testing.T) {
	provider := New(100)
	defer func() {
		if err := provider.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	logger := slog.New(provider)
	logger.Info("test integration message", "key", "value")

	ctx := context.Background()
	record, err := provider.Read(ctx)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
	if record == nil {
		t.Fatal("Read() returned nil record")
	}
	if record.Msg != "test integration message" {
		t.Errorf("Read() record.Msg = %v, want %v", record.Msg, "test integration message")
	}
}
