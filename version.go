package client

import "time"

func NewVersion() uint64 {
	return uint64(time.Now().UTC().UnixMilli())
}
