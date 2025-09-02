package snowflake

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var ErrInvalidSnowflake = errors.New("invalid snowflake")

func ParseSnowflake(value int64) Snowflake {
	return Snowflake(value)
}

func ParseStringSnowflake(str string) (Snowflake, error) {
	id, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return Snowflake(id), nil
}

type Snowflake int64

// Gets the node id of the Snowflake
func (s Snowflake) GetNodeId() (nodeId int64) {
	return (int64(s) >> snowflakeNodeIdShift) & ((1 << snowflakeNodeIdBits) - 1)
}

// Gets the raw timestamp (in unix millis) of the Snowflake
// You will have to manually add the timestamp offset
// to this value.
func (s Snowflake) GetTimestampRaw() (unixMillis int64) {
	return int64(s >> snowflakeTimestampShift)
}

// Gets the sequence number of the Snowflake.
func (s Snowflake) GetSeq() (seq int64) {
	return int64(s) & snowflakeSequenceMask
}

func (s Snowflake) String() string {
	return strconv.FormatInt(int64(s), 10)
}

func (s Snowflake) Value() (driver.Value, error) {
	return int64(s), nil
}

func (s *Snowflake) Scan(src any) error {
	switch v := src.(type) {
	case int64:
		*s = Snowflake(v)
		return nil
	case []byte:
		id, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}

		*s = Snowflake(id)
		return nil
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}

		*s = Snowflake(id)
		return nil
	default:
		return errors.New("incmopatiable type for Snowflake")
	}
}

func (s Snowflake) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(s), 10))
}

func (s *Snowflake) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("invalid snowflake JSON string: %w", err)
	}

	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid snowflake ID: %w", err)
	}

	*s = Snowflake(id)
	return nil
}
