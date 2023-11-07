package bitcoin

import (
	"math"
)

func ComputeDifficulty(hash []byte) float64 {
	var n float64 = 0

	for i := 0; i < len(hash); i++ {
		n = n + float64(hash[i])
		n = n * math.Pow(2, 8) //TODO: slow
	}

	return n
}

func MakeDifficulty(level uint64) []byte {
	difficulty := make([]byte, 32)
	for i := 0; i < 32; i++ {
		difficulty[i] = 255
	}
	for z := uint64(0); z < level; z++ {
		p := z / 8
		q := 7 - z%8
		difficulty[p] = difficulty[p] ^ (1 << q)
	}
	return difficulty
}

func AdjustDifficulty(state *State, blocksPerDifficulty uint64, blockInterval uint64) {
	if (state.LastBlockId+1)%(blocksPerDifficulty+1) == 0 {
		avgInterval := state.TotalInterval / (blocksPerDifficulty)
		state.Difficulty = state.Difficulty * float64((avgInterval / blockInterval))
	}
}
