package model

import (
	"time"
)

type Snapshot struct {
	Time time.Time
	JSON []byte
}
