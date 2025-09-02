# snowflake

high performance Snowflake library
built in `sql/driver` and json support

## installation
```bash
go get github.com/acelikesghosts/snowflake
```

## usage
```go
package main

import (
    "log"
    "github.com/acelikesghosts/snowflake"
)

// example: Twitter's (this also exists at snowflake.TwitterEpoch)
const myTimestampOffset = 1288834974657

func main() {
    nodeId := 1
    epochInMillis := snowflake.TwitterEpoch
    
    generator, _ := snowflake.NewGenerator(nodeId, epochInMillis)

    snowflake := generator.MustGenerate()
    log.Printf("snowflake: %s", snowflake.String())

    // unix millis
    timestamp := snowflake.GetTimestampRaw() + myTimestampOffset
    seq := snowflake.GetSeq()
    nodeId := snowflake.GetNodeId()

    log.Printf("timestamp (unix millis): %d\nseq: %d\nnode id: %d", timestamp, seq, nodeId)
}
```