# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v1.2.0] — 2026-04-12

### Summary

Phase 4S security hardening, documentation accuracy audit, and removal of
aspirational code that was never implemented. Clean, audit-ready release.

### Added

- **slogprovider package** (`github.com/agilira/iris/slogprovider`): bridges
  Go standard `log/slog` into the Iris pipeline. Implements both `slog.Handler`
  and `iris.SyncReader` simultaneously, allowing Metis and any slog-based service
  to adopt the Iris transport with zero call-site changes.

- **Phase 4S security test suite** — five new `security_test.go` files covering
  nine CWEs across all critical packages:
  - `autoscaling_security_test.go` — CWE-362 race conditions, nil-safety
  - `encoder_json_security_test.go` — CWE-116 (log injection), CWE-20 (input validation)
  - `encoder_text_security_test.go` — CWE-116, CWE-20, Unicode direction overrides
  - `internal/zephyroslite/security_test.go` — ring-buffer overflow, concurrent writes
  - `sampler_security_test.go` — CWE-400 resource exhaustion, race detection

- **CWE-400 ceiling on Config.Capacity** — `maxCapacity = 1<<24` constant added
  to `config.go`. `Validate()` now rejects oversized capacity configurations.

### Fixed

- **CWE-116 in JSON encoder** — field keys now pass through `quoteString()`.
  Previously only values were escaped; a malicious key could inject arbitrary
  JSON structure.

- **New() skipping Validate()** — `iris.New()` now calls `Validate()` after
  `buildSmartConfig()`. Previously, smart-config defaults could produce an
  invalid configuration that silently bypassed all validation checks.

- **IsValidLevel bounds** — `IsValidLevel` now correctly accepts the full range
  `Debug(-1)` through `Fatal(5)`. The previous bound capped at `Error(2)`,
  incorrectly rejecting `DPanic`, `Panic`, and `Fatal` as filter levels.

- **test_helpers.go excluded from production binary** — renamed to
  `test_helpers_test.go`. `TestConfig`/`TestConfigWithOutput`/`TestConfigSmall`
  renamed to `makeTestConfig*` variants to avoid confusion with Go's test runner.

### Removed

The following packages and files were aspirational specifications: they described
planned features but contained no working implementation. They were removed to
prevent misleading consumers of the library.

- `otel/` sub-module — OpenTelemetry integration (planned, never implemented)
- `lethebridge/` package — Lethe ring-buffer bridge (planned, never implemented)
- `magic.go` — Magic API layer (superseded by Smart API in `iris.go`)
- `config_loader.go` — Multi-source config loader with hot-reload (never working)
- `dynamic_config_test.go` — Tests for the above
- `integration_test.go` — Integration tests targeting removed aspirational features
- `examples/` directory — Examples for non-existent features
- `AUDIT_ENTERPRISE_TELEMETRY.md` — Aspirational enterprise telemetry doc
- `docs/OPENTELEMETRY.md` — No code exists
- `docs/HOT_RELOAD.md` — No code exists
- `docs/LOKI_INTEGRATION.md` — No code, no writer implementation
- `docs/CONFIGURATION_LOADING.md` — No code

### Changed

- **doc.go** — Complete rewrite. Removed all aspirational claims (OpenTelemetry,
  hot-reload, Console encoder, `logger.Dropped()`). Updated benchmarks to
  measured values. Corrected API references (`iris.Error`, `logger.Stats()`).

- **README.md** — Complete rewrite with fresh benchmark data measured on
  AMD Ryzen 5 7520U / linux/amd64 / Go 1.24.5. Removed references to
  non-existent sibling repos (`iris-writer-loki`, `iris-writer-datadog`).
  All examples use only real, implemented API.

- **docs/AUTOSCALING_ARCHITECTURE.md** — Rewritten to document the actual
  `AdaptiveLogger` / `ScalerConfig` / `AdaptiveStats` API. Removed references
  to the aspirational `AutoScalingLogger`, `AutoScalingConfig`, `AutoScalingMetrics`.

- **docs/ENCODERS.md** — Removed Console Encoder and Binary Encoder sections
  (neither is implemented). Only JSON and Text encoders documented.

- **docs/READERLOGGER_INTEGRATION.md** — Removed `iris.WithOTel()` and
  `iris.NewLokiWriter()` references (do not exist). Accurate `SyncReader` /
  `NewReaderLogger` API documented.

- **docs/CONTEXT_INTEGRATION.md** — Removed `WithRequestID()`, `WithUserID()`,
  `WithTraceID()` convenience method references (removed from codebase).
  Accurate `WithContext()` / `WithExtractor()` / `WithKeys()` / `WithKey()` API.

- **docs/SECURE_BY_DESIGN.md** — Removed aspirational sections: Custom
  `SecurityPolicy` struct, `stats.RedactedFields`, future ML-based anomaly
  detection, encryption-at-rest, digital signatures.

---

## [v1.1.1] — 2025-09-10

Security fixes and stability improvements. See git log for details.

---

## [v1.1.0] — 2025-08-22

Smart API, Auto-Scaling, Token Bucket Sampler, five Idle Strategies,
two Backpressure policies, SyncReader/SyncWriter interfaces, security
hardening (CWE-116 text encoder, CWE-362 race tests).

---

## [v1.0.0] — 2025-08-01

Initial public release. Ring-buffer architecture, JSON and Text encoders,
structured fields, context integration.

[v1.2.0]: https://github.com/agilira/iris/compare/v1.1.1...v1.2.0
[v1.1.1]: https://github.com/agilira/iris/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/agilira/iris/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/agilira/iris/releases/tag/v1.0.0
