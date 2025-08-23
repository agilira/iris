# IRIS Benchmark Notes

## Benchmark Design Philosophy

The benchmarks in IRIS are designed to be as fair and realistic as possible, following these principles:

1. **Zero-allocation path** - Core benchmarks measure the zero-allocation hot path for logging
2. **API helper costs separated** - Helper methods may cause allocations but these are separate from core logging
3. **Real-world usage patterns** - Benchmarks simulate both optimal and realistic usage patterns
4. **Zap compatibility** - Benchmarks mirror Zap's benchmarking approach for direct comparison

## Expected Allocations

Some API methods in IRIS will cause allocations for convenience. This is documented here:

### Zero-allocation Methods
- Basic logging methods (`Debug`, `Info`, `Warn`, `Error`)
- Most field creation methods (`String`, `Int`, `Bool`, etc.)

### Methods That May Cause Allocations
- `With()` - Creates a new logger with fields
- `Errors()` - When passed a dynamically created error slice
- Runtime reflection methods
- Stacktrace capture

## Pre-allocation vs. Dynamic Allocation

Some benchmarks use pre-allocated objects to demonstrate optimal usage. This is realistic in many production scenarios where loggers and fields are reused. Both patterns are benchmarked:

- `Benchmark10Fields` - Uses pre-allocated error
- `Benchmark10FieldsWithNewError` - Creates a new error each time
- `Benchmark100Fields` - Uses pre-allocated fields
- `Benchmark100FieldsWithCreation` - Creates fields dynamically

## Output Handling

For real-world scenarios with output, see `BenchmarkRealProcessing`.

## Comparison with other libraries

Iris benchmarks were designed to be directly comparable with Other Major Libraries, measuring the same operations in the same way.

## Real-world Verification

The `BenchmarkRealProcessing` and `BenchmarkCompareProcessingSpeed` benchmarks verify that logs are actually being processed, not just discarded or queued without processing. These confirm that the ring buffer consumer is working correctly.
