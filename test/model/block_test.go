package model

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/model"
	"log"
	"testing"
)

func Test_Serialize(t *testing.T) {

}

func Test_Difficulty(t *testing.T) {
	hash, err := cryptography.Hash("hello")
	if err != nil {
		t.Fatalf("compute hash err: %v", err)
	}

	for z := 1; z <= 5; z++ {
		difficulty := model.ComputeDifficulty(model.MakeDifficulty(z))
		log.Printf("difficulty: %x", difficulty)

		for k := z; k < z+5; k++ {
			for i := 0; i < k; i++ {
				p := i / 8
				q := 7 - i%8
				hash[p] = (hash[p] | (1 << q)) ^ (1 << q)
			}
			t.Logf("hash: %x", hash)

			actual := model.ComputeDifficulty(hash)
			if actual > difficulty {
				t.Fatalf("compute difficulty error, %v should smaller then %v, but actally greater", actual, difficulty)
			}
		}
	}
}
