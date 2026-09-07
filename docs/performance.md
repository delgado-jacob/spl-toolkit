---
title: "Performance baseline"
layout: page
---

# Performance baseline

This page records observations for fixed supported fixtures. It does not define a pass/fail threshold or predict throughput for other queries.

The benchmark used the accepted code SHA `6b55f8902ff1a990aea8951cd78934b32c90660a`, Go 1.26.8, macOS 26.2 (Darwin 25.2.0, arm64), and an Apple M1 Max CPU. Each benchmark was repeated five times with allocation reporting:

```bash
go test -mod=readonly ./pkg/mapper -run '^$' -bench '^BenchmarkBaseline' -benchmem -count=5
```

Parse and discovery used `search src_ip=1 | stats count by src_ip`. Map and parallel map used the same query with `src_ip` mapped to `source_ip`. The parallel benchmark used Go's `RunParallel` with one preconfigured mapper.

| Operation | Repetitions | Median ns/op | Observed ns/op range | Median B/op | allocs/op |
|---|---:|---:|---:|---:|---:|
| Parse | 5 | 421,330 | 419,252–433,054 | 417,188 | 5,737 |
| Discover | 5 | 438,375 | 432,406–568,316 | 426,694 | 5,726 |
| Map | 5 | 877,025 | 874,781–884,145 | 851,313 | 11,436 |
| Parallel map | 5 | 590,953 | 566,453–729,308 | 867,958 | 11,439 |

Raw output and the command environment are stored in `build/evidence/final-benchmark.txt` in the local build evidence (SHA-256 `042712107db811f38317304abab214144ab28998cc83aaf9104b0ba42932fe87`). This later evidence-only documentation commit does not change the benchmark, mapper, parser, generated parser, or dependency inputs measured at the recorded SHA.
