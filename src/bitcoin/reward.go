package bitcoin

const (
	InitReward = 50
)

func ComputeReward(lastBlockId uint64, blocksPerRewrad uint64) uint64 {
	var reward uint64 = InitReward
	if lastBlockId > 0 {
		reward /= (lastBlockId+1)/(blocksPerRewrad+1) + 1
	}
	return reward
}
