# Benchmark Comparison: Zap vs Iris

## Results Summary

| Benchmark | Zap (ns/op) | Iris (ns/op) | Zap Allocs | Iris Allocs | Performance Ratio |
|-----------|-------------|--------------|------------|-------------|-------------------|
| NoContext | 91.27 | 300,155 | 0 | 0 | **Zap 3,289x faster** |
| BoolField | 135.5 | 300,157 | 1 | 0 | **Zap 2,214x faster** |
| ByteStringField | 164.4 | 300,128 | 2 | 1 | **Zap 1,826x faster** |
| Float64Field | 156.2 | 300,127 | 1 | 1 | **Zap 1,921x faster** |
| IntField | 159.4 | 300,121 | 1 | 0 | **Zap 1,883x faster** |
| Int64Field | 142.8 | 300,186 | 1 | 0 | **Zap 2,102x faster** |
| StringField | 148.0 | 300,159 | 1 | 0 | **Zap 2,028x faster** |
| StringerField | 150.9 | 300,173 | 1 | 0 | **Zap 1,989x faster** |
| TimeField | 158.3 | 300,185 | 1 | 1 | **Zap 1,896x faster** |
| DurationField | 170.6 | 300,190 | 1 | 1 | **Zap 1,759x faster** |
| ErrorField | 155.9 | 300,171 | 1 | 0 | **Zap 1,925x faster** |
| ErrorsField | 276.2 | 300,307 | 2 | 7 | **Zap 1,087x faster** |
| ObjectField | 193.6 | 300,604 | 1 | 6 | **Zap 1,553x faster** |
| 10Fields | 451.0 | 300,232 | 1 | 0 | **Iris 1.5x faster** |
| 100Fields | 4,666 | 301,210 | 5 | 0 | **Iris 15.5x faster** |

## Key Observations

### Suspicious Pattern in Iris Results
- All Iris benchmarks show ~300µs timing (300,000ns)
- This suggests a fixed overhead/delay in the benchmark setup
- The consistent timing across different operations is not natural

### Where Iris Performs Better
- **10Fields**: Iris shows 300µs vs Zap's 451ns - but this seems inconsistent with the pattern
- **100Fields**: Iris shows 301µs vs Zap's 4.6µs - significant improvement in multi-field scenarios

### Disabled Benchmarks (Working Correctly)
- **DisabledWithoutFields**: Iris 9.4ns vs expected ~0ns (excellent)
- **DisabledAccumulatedContext**: Iris 24.2ns (reasonable)
- **DisabledAddingFields**: Iris 116ns with 640B/1alloc (needs investigation)

## Analysis
The consistent ~300µs timing in Iris suggests there might be:
1. A sleep/wait in the benchmark setup
2. Synchronous I/O that shouldn't be happening
3. A configuration issue causing artificial delays
4. The MPSC queue batching mechanism introducing fixed delays

The fact that 10Fields and 100Fields show the same ~300µs timing as single fields indicates the delay is not related to field processing but to the overall logging mechanism.
