package snowflake

import (
	"sync"
	"testing"
	"time"
)

func TestSnowflakeTimestamp(t *testing.T) {
	const (
		nodeID         = 1
		customEpoch    = 1288834974657 // Twitter epoch
		timestampShift = 22
	)

	gen, err := NewGenerator(nodeID, customEpoch)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	sf := gen.MustGenerate()

	sfInt := int64(sf)
	timestampPart := sfInt >> timestampShift

	actualTimestamp := timestampPart + (customEpoch)

	now := time.Now().UnixMilli()

	if diff := absDiff(now, actualTimestamp); diff > 5 {
		t.Errorf("Timestamp mismatch: expected ~%d, got %d (diff=%dms)", now, actualTimestamp, diff)
	}
}

func absDiff(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}

func TestSnowflakeGenerateBasic(t *testing.T) {
	gen, _ := NewGenerator(1, DiscordEpoch)
	id := gen.MustGenerate()

	if int64(id) == 0 {
		t.Fatal("Generated snowflake ID should not be zero")
	}

	ts := int64(id) >> snowflakeTimestampShift
	node := (int64(id) >> snowflakeNodeIdShift) & ((1 << snowflakeNodeIdBits) - 1)
	seq := int64(id) & snowflakeSequenceMask

	if node != 1 {
		t.Errorf("Expected node ID 1, got %d", node)
	}

	if seq != 0 {
		t.Errorf("Expected sequence 0 on first ID, got %d", seq)
	}

	if ts == 0 {
		t.Errorf("Timestamp part should not be zero")
	}
}

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
	_, err := NewGenerator((1 << snowflakeNodeIdBits), TwitterEpoch)
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

func TestSnowflakeGenerate_Concurrent(t *testing.T) {
	gen, _ := NewGenerator(1, DiscordEpoch)

	const concurrency = 100
	const idsPerGoroutine = 1000

	idsCh := make(chan Snowflake, concurrency*idsPerGoroutine)
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := gen.Generate()
				if err != nil {
					errCh <- err
					return
				}
				idsCh <- id
			}
			errCh <- nil
		}()
	}

	for i := 0; i < concurrency; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("unexpected error during generation: %v", err)
		}
	}
	close(idsCh)

	seen := make(map[Snowflake]struct{}, concurrency*idsPerGoroutine)
	for id := range idsCh {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate ID generated: %d", id)
		}
		seen[id] = struct{}{}
	}
}

func BenchmarkSnowflakeGenerate(b *testing.B) {
	gen, _ := NewGenerator(1, DiscordEpoch)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = gen.MustGenerate()
	}
}

func BenchmarkSnowflakeGenerate_Concurrent(b *testing.B) {
	gen, _ := NewGenerator(1, DiscordEpoch)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = gen.MustGenerate()
		}
	})
}
