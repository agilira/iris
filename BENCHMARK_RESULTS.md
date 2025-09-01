# Benchmark Comparison: Iris vs Zap (Complete Analysis)

## Results Summary

| Benchmark | Zap (ns/op) | Iris (ns/op) | Zap Allocs/B | Iris Allocs/B | Performance Gain |
|-----------|-------------|--------------|--------------|---------------|------------------|
| **NoContext** | 91.27 | **28.10** | 0 B/0 | 0 B/0 | **🚀 Iris 3.2x faster** |
| **BoolField** | 135.5 | **24.58** | 64 B/1 | 0 B/0 | **🚀 Iris 5.5x faster** |
| **ByteStringField** | 164.4 | **20.93** | 88 B/2 | 3 B/1 | **🚀 Iris 7.9x faster** |
| **Float64Field** | 156.2 | **24.72** | 64 B/1 | 0 B/0 | **🚀 Iris 6.3x faster** |
| **IntField** | 159.4 | **24.92** | 64 B/1 | 0 B/0 | **🚀 Iris 6.4x faster** |
| **Int64Field** | 142.8 | **24.53** | 64 B/1 | 0 B/0 | **🚀 Iris 5.8x faster** |
| **StringField** | 148.0 | **24.65** | 64 B/1 | 0 B/0 | **🚀 Iris 6.0x faster** |
| **StringerField** | 150.9 | **24.40** | 64 B/1 | 0 B/0 | **🚀 Iris 6.2x faster** |
| **TimeField** | 158.3 | **24.56** | 64 B/1 | 0 B/0 | **🚀 Iris 6.4x faster** |
| **DurationField** | 170.6 | **24.66** | 64 B/1 | 0 B/0 | **🚀 Iris 6.9x faster** |
| **ErrorField** | 155.9 | **24.45** | 64 B/1 | 0 B/0 | **🚀 Iris 6.4x faster** |
| **ErrorsField** | 276.2 | **149.8** | 88 B/2 | 184 B/7 | **🚀 Iris 1.8x faster** |
| **ObjectField** | 193.6 | **250.7** | 64 B/1 | 232 B/6 | ❌ Zap 1.3x faster |
| **10Fields** | 451.0 | **31.58** | 706 B/1 | 0 B/0 | **🚀 Iris 14.3x faster** |
| **100Fields** | 4,666 | **2,031** | 1297 B/5 | 5376 B/1 | **🚀 Iris 2.3x faster** |

## Key Performance Insights

### 🏆 **Iris Dominance**
- **13 out of 15 benchmarks**: Iris significantly outperforms Zap
- **Average speedup**: 5-7x faster for single field operations
- **Zero allocations**: Most Iris benchmarks have 0 allocations vs Zap's 1 allocation per operation

### 🎯 **Standout Performances**
- **10Fields**: Iris is **14.3x faster** (31.58ns vs 451ns) with **0 allocations** vs Zap's 706B
- **ByteStringField**: Iris is **7.9x faster** (20.93ns vs 164.4ns) 
- **DurationField**: Iris is **6.9x faster** (24.66ns vs 170.6ns)

### ⚡ **Multi-Field Scaling**
- **100Fields**: Iris is **2.3x faster** even with higher memory usage (5376B vs 1297B)
- Iris maintains performance advantage even on complex operations

### 🔴 **Where Zap Wins**
- **ObjectField**: Zap is 1.3x faster (193.6ns vs 250.7ns) 
  - This is due to Iris using string conversion `fmt.Sprintf("%+v")` vs Zap's native object marshaling

### 📊 **Memory Efficiency**
- **Iris**: Most operations are zero-allocation 
- **Zap**: Consistent 64B/1alloc pattern for single fields
- **Iris advantage**: Dramatically lower memory pressure in production

## Technical Analysis

### Why Iris Is Faster
1. **Zero-allocation field processing**: Iris's Field API avoids allocations that zap incurs
2. **Optimized encoding path**: Direct encoding without intermediate allocations
3. **Efficient ring buffer**: Even in sync mode, provides optimized record handling

### Where Iris Could Improve
1. **Object serialization**: Implement native object marshaling instead of `fmt.Sprintf`
2. **Complex operations**: Optimize string handling in ErrorsField

## Conclusion
**Iris delivers 3-14x performance improvements** over Zap in most scenarios while maintaining **zero allocations** for core logging operations. This represents a significant advancement in Go logging performance.
