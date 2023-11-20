package bitcoin

import (
	"Bitcoin/src/infra"
	"Bitcoin/src/model"
	"sync"
)

const (
	InitReward = 50
)

//TODO: test case for this class

type State struct {
	lock          sync.Mutex
	totalInterval uint64
	difficulty    float64
	lastBlock     *model.Block
}

func NewState(initDifficultyLevel uint64) *State {
	return &State{
		difficulty: infra.ComputeDifficulty(infra.MakeDifficulty(initDifficultyLevel)), //TODO: how to set when server restart?
		lock:       sync.Mutex{},
	}
}

func (state *State) Update(lastBlock *model.Block) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.totalInterval += uint64(lastBlock.Time.Sub(state.lastBlock.Time).Milliseconds())
	state.lastBlock = lastBlock
}

func (state *State) Get(blocksPerDifficulty, blocksPerRewrad, expectBlockInterval uint64) (uint64, float64) {
	state.lock.Lock()
	defer state.lock.Unlock()

	if state.lastBlock.Number%blocksPerDifficulty == 0 {
		avgInterval := state.totalInterval / (blocksPerDifficulty)
		state.difficulty = state.difficulty * float64((avgInterval / expectBlockInterval))
		state.totalInterval = 0
	}
	reward := InitReward / (state.lastBlock.Number/blocksPerRewrad + 1)

	return reward, state.difficulty
}

func (state *State) GetLastBlock() *model.Block {
	state.lock.Lock()
	defer state.lock.Unlock()
	return state.lastBlock
}
