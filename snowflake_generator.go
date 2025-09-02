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
	// 41 instead of 42 due to using a signed int
	snowflakeTimestampBits = 41
	snowflakeNodeIdBits    = 10
	snowflakeSequenceBits  = 12

	snowflakeNodeIdShift    = snowflakeSequenceBits
	snowflakeTimestampShift = snowflakeSequenceBits + snowflakeNodeIdBits
	snowflakeSequenceMask   = (1 << snowflakeSequenceBits) - 1
)

type SnowflakeGenerator struct {
	// timsetamp offset
	tsepoch int64
	// node id
	nid int64
	// last timestamp
	lts int64
	// start time
	st time.Time
}

var (
	ErrInvalidNodeId           error = errors.New("node is exceeds the maximum of 10 bits")
	ErrGenerationClockRollback error = errors.New("clock went backwards")
)

func NewGenerator(nodeId, timestampEpoch int64) (*SnowflakeGenerator, error) {
	if nodeId >= (1<<snowflakeNodeIdBits)-1 {
		return nil, ErrInvalidNodeId
	}

	return &SnowflakeGenerator{
		st:      time.Now(),
		nid:     nodeId,
		tsepoch: timestampEpoch,
	}, nil
}

func (s *SnowflakeGenerator) MustGenerate() Snowflake {
	snowflake, err := s.Generate()
	if err != nil {
		panic(err)
	}

	return snowflake
}

func (s *SnowflakeGenerator) Generate() (Snowflake, error) {
	for {
		elapsed := time.Since(s.st).Milliseconds()
		now := elapsed

		old := atomic.LoadInt64(&s.lts)
		oldTs := old >> snowflakeSequenceBits
		oldSeq := old & snowflakeSequenceMask

		var newTs, newSeq int64

		switch {
		case now > oldTs:
			newTs = now
			newSeq = 0

		case now == oldTs:
			if oldSeq == snowflakeSequenceMask {
				// seq overflow, wait for next millisecond
				time.Sleep(time.Millisecond)
				continue
			}

			newTs = now
			newSeq = oldSeq + 1

		case now < oldTs:
			return Snowflake(0), ErrGenerationClockRollback
		}

		packed := (newTs << snowflakeSequenceBits) | newSeq
		if atomic.CompareAndSwapInt64(&s.lts, old, packed) {
			finalTs := newTs + (s.tsepoch / 1e3)
			return Snowflake(
				(finalTs << snowflakeTimestampShift) |
					(s.nid << snowflakeNodeIdShift) |
					newSeq), nil
		}
	}
}
