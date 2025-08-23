# IRIS Changelog

## [1.5.0] - 2025-08-23 - CRITICAL SECURITY UPDATE

### 🚨 CRITICAL BUG FIX
- **Fixed data loss vulnerability in `Sync()` method** - Previous versions had a critical bug where `Sync()` was effectively a no-op, potentially causing log loss during application shutdown
- **Impact**: Applications using `defer logger.Sync()` patterns were at risk of losing critical log records
- **Action Required**: All users must upgrade immediately and review shutdown procedures

### ✅ What's Fixed
- `Sync()` now properly waits for all buffered records to be written to output
- Added 5-second timeout protection to prevent indefinite blocking
- Enhanced error reporting with detailed flush status information
- Fixed deadlock potential in custom WriteSyncer implementations

### 📚 Documentation Added
- **[Sync Integration Guide](docs/SYNC_INTEGRATION_GUIDE.md)** - Comprehensive patterns and best practices
- **[Migration Guide](docs/SYNC_MIGRATION_GUIDE.md)** - Critical update and upgrade instructions
- Added troubleshooting section for common integration issues
- Performance impact analysis and optimization guidelines

### 🔧 Technical Details
- ZephyrosLight `Flush()` method now returns error and waits for completion
- Ring buffer sync operations properly track processing completion
- Enhanced integration test suite covering slow I/O scenarios
- Fixed mutex deadlock patterns in test infrastructure

### ⚠️ Breaking Changes
- `Sync()` may now take longer to complete (this is correct behavior)
- Custom WriteSyncer implementations may need review for deadlock prevention
- Error handling for `Sync()` is now critical for data integrity

### 🚀 Migration
```go
// ✅ REQUIRED: Always check Sync() errors
if err := logger.Sync(); err != nil {
    fmt.Fprintf(os.Stderr, "Critical: failed to persist logs: %v\n", err)
}
logger.Close()
```

### 📊 Performance Impact
- Empty buffer sync: < 1μs (negligible overhead)
- Normal operation: 100μs - 10ms (expected for data integrity)
- Slow I/O: Up to 5s timeout (configurable in future versions)

---

## Previous Versions

### [1.4.x] - 2025-08-xx
- Context integration features
- Configuration management
- Security enhancements
- Sugar API documentation

### [1.3.x] - 2025-07-xx  
- Enterprise security features
- Data masking capabilities
- Log injection protection

### [1.2.x] - 2025-06-xx
- Performance optimizations
- ZephyrosLight integration
- Zero-allocation improvements

---

## Upgrade Priority

| Version | Risk Level | Action |
|---------|------------|---------|
| < 1.5.0 | 🔴 **CRITICAL** | **Upgrade immediately** |
| 1.5.0+ | 🟢 Safe | Regular maintenance |

For detailed migration instructions, see [docs/SYNC_MIGRATION_GUIDE.md](docs/SYNC_MIGRATION_GUIDE.md).
