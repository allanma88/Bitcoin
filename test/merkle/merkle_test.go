package merkle

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/merkle"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func Test_Merkle_Marshal_Succeed(t *testing.T) {
	n := 5
	vals := make([]string, n)
	for i := 0; i < n; i++ {
		vals[i] = fmt.Sprintf("Hello%d", i)
	}

	tree, err := merkle.BuildTree[string](vals)
	if err != nil {
		t.Fatalf("build merkle tree error: %s", err)
	}

	data, err := json.Marshal(tree)
	if err != nil {
		t.Fatalf("hash merkle tree error: %s", err)
	}

	t.Logf("json: %s", string(data))

	var newtree merkle.MerkleTree[string]
	err = json.Unmarshal(data, &newtree)
	if err != nil {
		t.Fatalf("unmarshal merkle tree error: %s", err)
	}

	print(&newtree, t)
}

func Test_Merkle_Marshal_Duplicate_Failed(t *testing.T) {
	n := 7
	vals := make([]string, n)
	for i := 0; i < n; i++ {
		vals[i] = "Hello"
	}

	tree, err := merkle.BuildTree[string](vals)
	if err != nil {
		t.Fatalf("build merkle tree error: %s", err)
	}

	data, err := json.Marshal(tree)
	if err != nil {
		t.Fatalf("hash merkle tree error: %s", err)
	}

	t.Logf("json: %s", string(data))

	var newtree merkle.MerkleTree[string]
	err = json.Unmarshal(data, &newtree)

	nodeUnmarshalError := err.(merkle.NodeUnmarshalError)
	if nodeUnmarshalError.Err != merkle.ErrMtDuplicateHash {
		t.Fatalf("expect error: %s, actual err: %v", merkle.ErrMtDuplicateHash, err)
	}
}

func Test_Merkle_Marshal_Batch_Succeed(t *testing.T) {
	for n := 2; n < 50; n++ {
		vals := make([]string, n)
		for i := 0; i < n; i++ {
			vals[i] = fmt.Sprintf("Hello%d", i)
		}

		tree, err := merkle.BuildTree[string](vals)
		if err != nil {
			t.Fatalf("build merkle tree error: %s", err)
		}

		data, err := json.Marshal(tree)
		if err != nil {
			t.Fatalf("hash merkle tree error: %s", err)
		}

		var newtree merkle.MerkleTree[string]
		err = json.Unmarshal(data, &newtree)
		if err != nil {
			t.Fatalf("unmarshal merkle tree error: %s", err)
		}
	}
}

func Test_Merkle_Unmarshal_Failed(t *testing.T) {
	files := make(map[string]string)
	files["merkle_unmarshal_failed_node_size.json"] = merkle.ErrMtNodeSize
	files["merkle_unmarshal_failed_few_rows.json"] = merkle.ErrMtFewRows
	files["merkle_unmarshal_failed_few_parent_hash.json"] = merkle.ErrMtFewParentHash
	files["merkle_unmarshal_failed_no_parent_hash.json"] = merkle.ErrMtNoParentHash
	files["merkle_unmarshal_failed_duplicate_hash.json"] = merkle.ErrMtDuplicateHash

	for file, expect := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read json file error: %s", err)
		}

		var tree merkle.MerkleTree[string]
		err = json.Unmarshal(data, &tree)

		rowUnmarshalError, ok := err.(merkle.RowUnmarshalError)
		if ok && rowUnmarshalError.Err != expect {
			t.Fatalf("%s unmarshal merkle tree, expect: %v, actual: %v", file, expect, rowUnmarshalError)
		}

		nodeUnmarshalError, ok := err.(merkle.NodeUnmarshalError)
		if ok && nodeUnmarshalError.Err != expect {
			t.Fatalf("%s unmarshal merkle tree, expect: %v, actual: %v", file, expect, nodeUnmarshalError)
		}
	}
}

func Test_Merkle_Hash(t *testing.T) {
	vals := []string{"Hello1", "Hello2", "Hello3", "Hello4", "Hello5", "Hello6", "Hello7"}

	tree, err := merkle.BuildTree[string](vals)
	if err != nil {
		t.Fatalf("build merkle tree error: %s", err)
	}

	hash, err := cryptography.Hash(tree)
	if err != nil {
		t.Fatalf("hash merkle tree error: %s", err)
	}

	t.Logf("hash: %x", hash)
}

func Test_Merkle_Validate_Success(t *testing.T) {
	for n := 2; n < 100; n++ {
		vals := make([]string, n)
		for i := 0; i < n; i++ {
			vals[i] = fmt.Sprintf("Hello%d", i)
		}

		tree, err := merkle.BuildTree[string](vals)
		if err != nil {
			t.Fatalf("build merkle tree error: %s", err)
		}

		data, err := json.Marshal(tree)
		if err != nil {
			t.Fatalf("marshal merkle tree error: %s", err)
		}
		t.Logf("json: %s", string(data))

		valid, err := tree.Validate()
		if err != nil {
			t.Fatalf("validate merkle tree error: %s", err)
		}
		if !valid {
			t.Fatalf("validate merkle tree should succeed, but failed")
		}
	}
}

func Test_Merkle_Validate_Fail(t *testing.T) {
	vals := []string{"Hello1", "Hello2", "Hello3", "Hello4", "Hello5", "Hello6", "Hello7"}

	tree, err := merkle.BuildTree[string](vals)
	if err != nil {
		t.Fatalf("build merkle tree error: %s", err)
	}
	tree.Table[0][0].Hash[0] = 1

	valid, err := tree.Validate()
	if err != nil {
		t.Fatalf("validate merkle tree error: %s", err)
	}
	if valid {
		t.Fatal("validate merkle tree should failed, but succeed")
	}
}

func Test_Merkle_Validate_Success_From_Json(t *testing.T) {
	data, err := os.ReadFile("merkle_success.json")
	if err != nil {
		t.Fatalf("read json file error: %s", err)
	}

	var tree merkle.MerkleTree[string]
	err = json.Unmarshal(data, &tree)
	if err != nil {
		t.Fatalf("unmarshal merkle tree error: %s", err)
	}

	valid, err := tree.Validate()
	if err != nil {
		t.Fatalf("validate merkle tree error: %s", err)
	}
	if !valid {
		t.Fatalf("the valid merkle tree validate failed")
	}
}

func Test_Merkle_Validate_Fail_From_Json(t *testing.T) {
	data, err := os.ReadFile("merkle_validate_failed.json")
	if err != nil {
		t.Fatalf("read json file error: %s", err)
	}

	var tree merkle.MerkleTree[string]
	err = json.Unmarshal(data, &tree)
	if err != nil {
		t.Fatalf("unmarshal merkle tree error: %s", err)
	}

	valid, err := tree.Validate()
	if err != nil {
		t.Fatalf("validate merkle tree error: %s", err)
	}
	if valid {
		t.Fatal("validate merkle tree should failed, but succeed")
	}
}

func Test_Merkle_Get(t *testing.T) {
	for n := 2; n < 100; n++ {
		vals := make([]string, n)
		hashs := make([][]byte, n)
		for i := 0; i < n; i++ {
			vals[i] = fmt.Sprintf("Hello%d", i)
			hash, err := cryptography.Hash(vals[i])
			if err != nil {
				t.Fatalf("hash %s error: %s", vals[i], err)
			}
			hashs[i] = hash
		}

		tree, err := merkle.BuildTree[string](vals)
		if err != nil {
			t.Fatalf("build merkle tree error: %s", err)
		}

		for i, hash := range hashs {
			val, err := tree.Get(hash)
			if err != nil {
				t.Fatalf("search merkle tree error: %s", err)
			}
			if val == "" {
				t.Fatalf("didn't find the hash %x in the merkle tree", hash)
			}
			if val != vals[i] {
				t.Fatalf("expect get: %s, actual: %s", vals[i], val)
			}
		}
	}
}

func print[T any](tree *merkle.MerkleTree[T], t *testing.T) {
	for r := 0; r < len(tree.Table); r++ {
		for c := 0; c < len(tree.Table[r]); c++ {
			t.Logf("%x", tree.Table[r][c].Hash)
		}
		t.Log("\n")
	}
}
