package mapper

import (
	"fmt"
	"sync"
	"testing"
)

func TestAtomicLoadMappingsPublication(t *testing.T) {
	m := New()
	first := []byte(`[{"source":"a","target":"x"},{"source":"b","target":"y"}]`)
	second := []byte(`[{"source":"a","target":"p"},{"source":"b","target":"q"}]`)
	if err := m.LoadMappings(first); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	errs := make(chan error, 9)
	var wg sync.WaitGroup
	wg.Add(9)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 100; i++ {
			data := second
			if i%2 == 1 {
				data = first
			}
			if err := m.LoadMappings(data); err != nil {
				errs <- err
				return
			}
		}
	}()
	for i := 0; i < 8; i++ {
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 100; j++ {
				got, err := m.MapQuery("search a=1 b=2")
				if err != nil {
					errs <- err
					return
				}
				if got != "search x=1 y=2" && got != "search p=1 q=2" {
					errs <- fmt.Errorf("observed partially published mappings: %q", got)
					return
				}
			}
		}()
	}

	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestConcurrentParserReuseKeepsErrorsOperationLocal(t *testing.T) {
	p := NewParser()
	start := make(chan struct{})
	errs := make(chan error, 200)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			if _, err := p.Parse("search host=web"); err != nil {
				errs <- fmt.Errorf("valid parse failed: %w", err)
			}
		}()
		go func() {
			defer wg.Done()
			<-start
			if _, err := p.Parse("|"); err == nil {
				errs <- fmt.Errorf("invalid parse succeeded")
			}
		}()
	}

	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}
