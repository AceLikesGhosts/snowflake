package snowflake

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestSnowflakeJSONMarshaling(t *testing.T) {
	gen, _ := NewGenerator(2, DiscordEpoch)
	id := gen.MustGenerate()

	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("Failed to marshal snowflake: %v", err)
	}

	var id2 Snowflake
	if err := json.Unmarshal(data, &id2); err != nil {
		t.Fatalf("Failed to unmarshal snowflake: %v", err)
	}

	if id != id2 {
		t.Errorf("Unmarshaled snowflake does not match original. Got %d, want %d", id2, id)
	}
}

func TestSnowflakeDatabaseValueScan(t *testing.T) {
	gen, _ := NewGenerator(3, DiscordEpoch)
	id := gen.MustGenerate()

	val, err := id.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}

	i64, ok := val.(int64)
	if !ok {
		t.Fatalf("Value() should return int64, got %T", val)
	}

	var s1 Snowflake
	if err := s1.Scan(i64); err != nil {
		t.Fatalf("Scan(int64) returned error: %v", err)
	}

	if s1 != id {
		t.Errorf("Scan(int64) result mismatch: got %d want %d", s1, id)
	}

	var s2 Snowflake

	str := []byte(strconv.FormatInt(int64(id), 10))
	if err := s2.Scan(str); err != nil {
		t.Fatalf("Scan([]byte) returned error: %v", err)
	}

	var s3 Snowflake
	if err := s3.Scan(strconv.FormatInt(int64(id), 10)); err != nil {
		t.Fatalf("Scan(string) returned error: %v", err)
	}

	var s4 Snowflake
	if err := s4.Scan(3.14); err == nil {
		t.Fatal("Expected error scanning unsupported type")
	}
}
