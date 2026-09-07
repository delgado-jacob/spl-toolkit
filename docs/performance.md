---
title: "Performance baseline"
layout: page
---

# Performance baseline

This page records observations for fixed supported fixtures. It does not define a pass/fail threshold or predict throughput for other queries.

The benchmark used code SHA `18ec33c2cd27e61d6295ce0a82e2aef56e88ef2e`, Go 1.26.8, macOS 26.2 (Darwin 25.2.0, arm64), and an Apple M1 Max CPU. Each benchmark was repeated five times with allocation reporting:

```bash
go test -mod=readonly ./pkg/mapper -run '^$' -bench '^BenchmarkBaseline' -benchmem -count=5
```

Parse and discovery used `search src_ip=1 | stats count by src_ip`. Map and parallel map used the same query with `src_ip` mapped to `source_ip`. The parallel benchmark used Go's `RunParallel` with one preconfigured mapper.

| Operation | Repetitions | Median ns/op | Observed ns/op range | Median B/op | allocs/op |
|---|---:|---:|---:|---:|---:|
| Parse | 5 | 417,269 | 415,985–432,739 | 417,188 | 5,737 |
| Discover | 5 | 435,448 | 434,376–446,061 | 426,849 | 5,726 |
| Map | 5 | 887,968 | 871,586–895,466 | 851,106 | 11,436 |
| Parallel map | 5 | 559,054 | 549,227–586,269 | 869,328 | 11,439 |

Raw output and the command environment are stored in `build/evidence/task-8-benchmark.txt` in the local build evidence. The later documentation and acceptance-harness commit does not change the benchmark, mapper, parser, generated parser, or dependency inputs measured at the recorded SHA.
