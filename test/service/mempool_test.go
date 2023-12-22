package service

import (
	"Bitcoin/src/service"
	"testing"
)

func Test_Load(t *testing.T) {
	mempool := service.NewMemPool(10)
	dir := "Bitcoin"
	if err := mempool.Load(dir); err != nil {
		t.Fatalf("load mempool from %v error: %v", dir, err)
	}
}
