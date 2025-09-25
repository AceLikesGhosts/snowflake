package snowflake

import (
	"errors"
	"sync/atomic"
	"time"
)

// presets for ease of use
// with other services.
var (
	TwitterEpoch int64 = 1288834974657
	DiscordEpoch int64 = 1420070400000
)

const (
	snowflakeTimestampBits = 41
	snowflakeNodeIdBits    = 10
	snowflakeSequenceBits  = 12

	snowflakeNodeIdShift    = snowflakeSequenceBits
	snowflakeTimestampShift = snowflakeSequenceBits + snowflakeNodeIdBits
	snowflakeSequenceMask   = (1 << snowflakeSequenceBits) - 1
)

var (
	ErrInvalidNodeId           = errors.New("node ID exceeds maximum of 10 bits")
	ErrGenerationClockRollback = errors.New("clock went backwards")
)

type SnowflakeGenerator struct {
	tsepoch    int64
	shiftedNid int64
	// packed: (timestamp << 12) | sequence
	state atomic.Int64
}

func NewGenerator(nodeId, timestampEpoch int64) (*SnowflakeGenerator, error) {
	if nodeId >= (1 << snowflakeNodeIdBits) {
		return nil, ErrInvalidNodeId
	}

	return &SnowflakeGenerator{
		tsepoch:    timestampEpoch,
		shiftedNid: nodeId << snowflakeNodeIdShift,
	}, nil
}

// MustGenerate panics on error
func (s *SnowflakeGenerator) MustGenerate() Snowflake {
	id, err := s.Generate()
	if err != nil {
		panic(err)
	}
	return id
}

// Generates a Snowflake.
func (s *SnowflakeGenerator) Generate() (Snowflake, error) {
	for {
		now := time.Now().UnixMilli()
		relativeTs := now - s.tsepoch

		prev := s.state.Load()
		prevTs := prev >> snowflakeSequenceBits
		prevSeq := prev & snowflakeSequenceMask

		var nextTs, nextSeq int64

		if relativeTs < prevTs {
			deadline := time.Now().Add(15 * time.Millisecond)
			for time.Now().Before(deadline) {
				now = time.Now().UnixMilli()
				relativeTs = now - s.tsepoch
				if relativeTs >= prevTs {
					break
				}
				time.Sleep(1 * time.Millisecond)
			}
			if relativeTs < prevTs {
				return 0, ErrGenerationClockRollback
			}
		}

		if relativeTs == prevTs {
			if prevSeq == snowflakeSequenceMask {
				time.Sleep(time.Millisecond)
				continue
			}
			nextTs = prevTs
			nextSeq = prevSeq + 1
		} else {
			nextTs = relativeTs
			nextSeq = 0
		}

		newState := (nextTs << snowflakeSequenceBits) | nextSeq

		if s.state.CompareAndSwap(prev, newState) {
			id := (nextTs << snowflakeTimestampShift) | s.shiftedNid | nextSeq
			return Snowflake(id), nil
		}
	}
}
