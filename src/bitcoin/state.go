package bitcoin

import (
	"sync"
	"time"
)

const (
	InitReward = 50
)

type State struct {
	lock            sync.Mutex
	totalInterval   uint64
	difficulty      float64
	lastBlockNumber uint64
	lastBlockTime   time.Time
}

func NewState(initDifficultyLevel uint64) *State {
	return &State{
		difficulty: ComputeDifficulty(MakeDifficulty(initDifficultyLevel)), //TODO: how to set when server restart?
		lock:       sync.Mutex{},
	}
}

func (state *State) Update(blockNumber uint64, blockTime time.Time) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.totalInterval += uint64(blockTime.Sub(state.lastBlockTime).Milliseconds())
	state.lastBlockNumber = blockNumber
	state.lastBlockTime = blockTime
}

func (state *State) Get(blocksPerDifficulty, blocksPerRewrad uint64, expectBlockInterval uint64) (uint64, uint64, float64) {
	state.lock.Lock()
	defer state.lock.Unlock()

	if state.lastBlockNumber%blocksPerDifficulty == 0 {
		avgInterval := state.totalInterval / (blocksPerDifficulty)
		state.difficulty = state.difficulty * float64((avgInterval / expectBlockInterval))
		state.totalInterval = 0
	}
	reward := InitReward / (state.lastBlockNumber/blocksPerRewrad + 1)

	return state.lastBlockNumber, reward, state.difficulty
}

func (state *State) GetLastBlockNumber() uint64 {
	state.lock.Lock()
	defer state.lock.Unlock()
	return state.lastBlockNumber
}
