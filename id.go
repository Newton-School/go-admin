package admin

import (
	"fmt"
	"strconv"
)

// IDCodec converts route IDs to repository IDs and back.
type IDCodec[ID comparable] interface {
	Parse(raw string) (ID, error)
	Format(id ID) string
}

type int64IDCodec struct{}

// Int64ID returns an ID codec for int64 primary keys.
func Int64ID() IDCodec[int64] {
	return int64IDCodec{}
}

func (int64IDCodec) Parse(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse int64 id %q: %w", raw, err)
	}
	return id, nil
}

func (int64IDCodec) Format(id int64) string {
	return strconv.FormatInt(id, 10)
}

type stringIDCodec struct{}

// StringID returns an ID codec for non-empty string primary keys.
func StringID() IDCodec[string] {
	return stringIDCodec{}
}

func (stringIDCodec) Parse(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("id cannot be empty")
	}
	return raw, nil
}

func (stringIDCodec) Format(id string) string {
	return id
}
