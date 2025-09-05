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

	// start time
	st time.Time
	// atomic int64, last timestamp used
	lts int64
	// atomic int64, seq number
	seq int64
}

func NewGenerator(nodeId, timestampEpoch int64) (*SnowflakeGenerator, error) {
	if nodeId >= (1 << snowflakeNodeIdBits) {
		return nil, ErrInvalidNodeId
	}

	return &SnowflakeGenerator{
		tsepoch:    timestampEpoch,
		shiftedNid: nodeId << snowflakeNodeIdShift,
		st:         time.Now(),
		seq:        0,
		lts:        0,
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

		lastTs := atomic.LoadInt64(&s.lts)
		seq := atomic.LoadInt64(&s.seq)

		if relativeTs < lastTs {
			return 0, ErrGenerationClockRollback
		}

		if relativeTs == lastTs {
			if seq == snowflakeSequenceMask {
				time.Sleep(time.Millisecond)
				continue
			}

			if !atomic.CompareAndSwapInt64(&s.seq, seq, seq+1) {
				continue
			}

			seq++
		} else {
			if !atomic.CompareAndSwapInt64(&s.lts, lastTs, relativeTs) {
				continue
			}
			atomic.StoreInt64(&s.seq, 0)
			seq = 0
		}

		id := (relativeTs << snowflakeTimestampShift) | s.shiftedNid | seq
		return Snowflake(id), nil
	}
}
