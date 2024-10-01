package resp

import (
	"fmt"

	"github.com/Drumstickz64/redis-go/data"
)

func EncodeSimpleString(s data.String) []byte {
	return []byte(data.String(fmt.Sprintf("+%s\r\n", s)))
}

func EncodeBulkString(s data.String) []byte {
	return []byte(data.String(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)))
}
