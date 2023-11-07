package bitcoin

import (
	"math"
	"time"
)

func ComputeDifficulty(hash []byte) float64 {
	var n float64 = 0

	for i := 0; i < len(hash); i++ {
		n = n + float64(hash[i])
		n = n * math.Pow(2, 8) //TODO: slow
	}

	return n
}

func MakeDifficulty(z int) []byte {
	difficulty := make([]byte, 32)
	for i := 0; i < 32; i++ {
		difficulty[i] = 255
	}
	for i := 0; i < z; i++ {
		p := i / 8
		q := 7 - i%8
		difficulty[p] = difficulty[p] ^ (1 << q)
	}
	return difficulty
}

func AdjustDifficulty(state *State, blocksPerDifficulty int, blockInterval time.Duration) {
	if (state.LastBlockId+1)%uint64(blocksPerDifficulty+1) == 0 {
		avgInterval := state.TotalInterval.Milliseconds() / int64(blocksPerDifficulty)
		state.Difficulty = state.Difficulty * float64((avgInterval / blockInterval.Milliseconds()))
	}
}
