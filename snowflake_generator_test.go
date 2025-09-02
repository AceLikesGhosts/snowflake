package snowflake

import (
	"sync"
	"testing"
	"time"
)

func TestNewSnowflakeGenerator_ValidNodeId(t *testing.T) {
	gen, err := NewGenerator(1, TwitterEpoch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected generator, got nil")
	}
}

func TestNewSnowflakeGenerator_InvalidNodeId(t *testing.T) {
	_, err := NewGenerator((1<<snowflakeNodeIdBits)-1, TwitterEpoch)
	if err != ErrInvalidNodeId {
		t.Fatalf("expected ErrInvalidNodeId, got %v", err)
	}
}

func TestSnowflakeGenerateConcurrent(t *testing.T) {
	gen, _ := NewGenerator(4, DiscordEpoch)
	const count = 10000

	var wg sync.WaitGroup
	wg.Add(count)

	idMap := sync.Map{}

	for range count {
		go func() {
			defer wg.Done()
			id := gen.MustGenerate()
			if _, loaded := idMap.LoadOrStore(id, true); loaded {
				t.Errorf("Duplicate ID generated: %d", id)
			}
		}()
	}

	wg.Wait()
}

func BenchmarkSnowflakeGenerate(b *testing.B) {
	gen, _ := NewGenerator(1, DiscordEpoch)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = gen.MustGenerate()
	}
}

func TestMustGenerate_PanicsOnError(t *testing.T) {
	gen := &SnowflakeGenerator{
		// future start time to simulate rollback
		st:      time.Now().Add(time.Hour),
		nid:     1,
		tsepoch: TwitterEpoch,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, but did not panic")
		}
	}()

	// should panic due to clock rollback
	_ = gen.MustGenerate()
}

func TestGenerate_ClockRollback(t *testing.T) {
	// simulate time rollback
	gen := &SnowflakeGenerator{
		st:      time.Now().Add(time.Hour),
		nid:     1,
		tsepoch: TwitterEpoch,
	}

	_, err := gen.Generate()
	if err != ErrGenerationClockRollback {
		t.Fatalf("expected ErrGenerationClockRollback, got %v", err)
	}
}
