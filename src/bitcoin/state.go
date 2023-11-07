package bitcoin

import "time"

type State struct {
	TotalInterval time.Duration
	Difficulty    float64
	LastBlockId   uint64
}
