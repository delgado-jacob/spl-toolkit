package mapper

import "testing"

const baselineBenchmarkQuery = "search src_ip=1 | stats count by src_ip"

func BenchmarkBaselineParse(b *testing.B) {
	p := NewParser()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := p.Parse(baselineBenchmarkQuery); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBaselineDiscover(b *testing.B) {
	m := New()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := m.DiscoverQuery(baselineBenchmarkQuery); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkMapper(b *testing.B) *Mapper {
	b.Helper()
	m := New()
	if err := m.LoadMappings([]byte(`[{"source":"src_ip","target":"source_ip"}]`)); err != nil {
		b.Fatal(err)
	}
	return m
}

func BenchmarkBaselineMap(b *testing.B) {
	m := benchmarkMapper(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := m.MapQuery(baselineBenchmarkQuery); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBaselineParallelMap(b *testing.B) {
	m := benchmarkMapper(b)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := m.MapQuery(baselineBenchmarkQuery); err != nil {
				b.Fatal(err)
			}
		}
	})
}
