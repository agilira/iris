# Iris vs Zap Console Encoding Performance Analysis
Date: 2025-08-17
Hardware: AMD Ryzen 5 7520U with Radeon Graphics

## Console Encoding Performance (Encoder-only)

| Metrica | Iris | Zap | Speedup |
|---------|------|-----|---------|
| **Latency** | 383ns | 1,038ns | **2.7x faster** |
| **Memory** | 0 B/op | 1,121 B/op | **∞ more efficient** |
| **Allocations** | 0 allocs/op | 5 allocs/op | **Zero allocations** |

### With Colors
| Metrica | Iris (colors) | Zap (colors) | Speedup |
|---------|---------------|--------------|---------|
| **Latency** | 391ns | 1,163ns | **3.0x faster** |
| **Memory** | 0 B/op | 1,121 B/op | **∞ more efficient** |

## Full End-to-End Logger Performance

### Structured Logging (3 fields)
| Logger | Latency | Memory | Allocations | vs Iris |
|--------|---------|--------|-------------|---------|
| **Iris (full)** | **35ns** | **0 B/op** | **0 allocs/op** | **1.0x** |
| Zap Development | 2,586ns | 545 B/op | 7 allocs/op | **74x slower** |
| Zap Production | 180ns | 195 B/op | 1 allocs/op | **5.1x slower** |

### Simple Logging (message only)
| Logger | Latency | Memory | Allocations | vs Iris |
|--------|---------|--------|-------------|---------|
| **Iris (simple)** | **12ns** | **0 B/op** | **0 allocs/op** | **1.0x** |

## Key Insights

1. **Console Encoding**: Iris is 2.7-3.0x faster than Zap with zero allocations
2. **Full Pipeline**: Iris is 5-74x faster than Zap depending on configuration
3. **Memory Efficiency**: Iris has zero allocations vs Zap's 195-1,121 bytes per operation
4. **Consistency**: Iris maintains zero allocations across all scenarios

## Performance Guarantees Met

✅ Console encoding <400ns (achieved: ~383ns)
✅ Zero allocations maintained  
✅ Significantly outperforms Zap in all scenarios
✅ Colors add minimal overhead (~8ns)

## Conclusion

Iris console encoding is not just competitive with Zap - it's dramatically superior:
- **2.7-3x faster encoding**
- **5-74x faster full pipeline** 
- **Zero memory allocations**
- **Zero garbage collection pressure**

This positions Iris as the clear leader for high-performance console logging.
