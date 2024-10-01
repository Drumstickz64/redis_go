package resp

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Drumstickz64/redis-go/data"
)

func Decode(str string) (data.Data, error) {
	decoder := Decoder{
		str:  str,
		curr: 0,
	}

	slog.Debug("decoding string", "string", str)

	return decoder.Decode()
}

type Decoder struct {
	str  string
	curr int
}

func (d *Decoder) Decode() (data.Data, error) {
	char := d.Advance()

	slog.Debug("now decoding section", "remaining", d.Remaining(), "matching", string(char))

	switch char {
	case '+':
		return d.DecodeSimpleString()
	case '-':
		return d.DecodeSimpleError()
	case ':':
		return d.DecodeInteger()
	case '$':
		return d.DecodeBulkString()
	case '*':
		return d.DecodeArray()
	default:
		return nil, fmt.Errorf("invalid character %c in position %d", rune(d.Prev()), d.curr-1)
	}
}

func (d *Decoder) DecodeSimpleString() (data.String, error) {
	slog.Debug("decoding simple string", "remaining", d.Remaining())

	str, err := d.ReadUntilTerminator()
	if err != nil {
		return "", fmt.Errorf("error while decoding simple string, %v", err)
	}

	slog.Debug("decoded simple string", "string", str)

	return data.String(str), nil
}

func (d *Decoder) DecodeSimpleError() (data.Error, error) {
	slog.Debug("decoding simple error", "remaining", d.Remaining())

	str, err := d.ReadUntilTerminator()
	if err != nil {
		return "", fmt.Errorf("error while decoding simple error, %v", err)
	}

	slog.Debug("decoded simple error", "string", str)

	return data.Error(str), nil
}

func (d *Decoder) DecodeInteger() (data.Integer, error) {
	slog.Debug("decoding integer", "remaining", d.Remaining())

	str, err := d.ReadUntilTerminator()
	if err != nil {
		return 0, fmt.Errorf("error while decoding integer: failed to read integer string, %v", err)
	}

	integer, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error while decoding integer: failed to parse integer string %s: %v", str, err)
	}

	slog.Debug("decoded integer", "integer", integer)

	return data.Integer(integer), nil
}

func (d *Decoder) DecodeBulkString() (data.Data, error) {
	slog.Debug("decoding bulk string", "remaining", d.Remaining())

	lenStr, err := d.ReadUntilTerminator()
	if err != nil {
		return "", fmt.Errorf("error while decoding bulk string: failed to read string length: %v", err)
	}

	len, err := strconv.ParseInt(lenStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("error while decoding bulk string: failed to parse string length: %v", err)
	}

	if len == -1 {
		return data.Null{}, nil
	} else if len < 0 {
		return nil, fmt.Errorf("error while decoding bulk string: length should be -1, 0, or a positive integer, but was %d", len)
	}

	str := d.str[d.curr : d.curr+int(len)]
	d.curr += int(len)

	if err := d.ExpectTerminator(); err != nil {
		return nil, fmt.Errorf("error while decoding bulk string: %v", err)
	}

	slog.Debug("parsed bulk string", "string", str)

	return data.String(str), nil
}

func (d *Decoder) DecodeArray() (data.Data, error) {
	slog.Debug("decoding array", "remaining", d.Remaining())

	lenStr, err := d.ReadUntilTerminator()
	if err != nil {
		return nil, fmt.Errorf("error while decoding array, failed to read array length: %v", err)
	}

	len, err := strconv.ParseInt(lenStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error while decoding array: failed to parse array length: %v", err)
	}

	if len == -1 {
		return data.Null{}, nil
	} else if len < 0 {
		return nil, fmt.Errorf("error while decoding array: length should be -1, 0, or a positive integer, but was %d", len)
	}

	arr := make([]data.Data, len)
	for i := range len {
		elem, err := d.Decode()
		if err != nil {
			return nil, fmt.Errorf("error while decoding array element %d: %v", i, err)
		}

		arr[i] = elem
	}

	slog.Debug("parsed array", "array", arr)

	return data.Array(arr), nil
}

func (d *Decoder) Advance() byte {
	d.curr++
	return d.str[d.curr-1]
}

func (d *Decoder) Peek() byte {
	return d.str[d.curr]
}

func (d *Decoder) Prev() byte {
	return d.str[d.curr-1]
}

func (d *Decoder) Remaining() string {
	return d.str[d.curr:]
}

func (d *Decoder) ReadUntilTerminator() (string, error) {
	start := d.curr

	for d.Advance() != '\r' {
	}

	if d.Advance() != '\n' {
		return "", fmt.Errorf("expected LF after CR: at position %d", d.curr)
	}

	return d.str[start : d.curr-2], nil
}

func (d *Decoder) ExpectTerminator() error {
	if d.str[d.curr:d.curr+2] != "\r\n" {
		return fmt.Errorf("error while decoding bulk string, expected a CRLF terminator at %d, found %c", d.curr, d.str[d.curr])
	}

	d.curr += 2

	return nil
}
