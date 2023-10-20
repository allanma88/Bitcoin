package merkle

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/merkle"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func Test_Merkle_Marshal_Single_Succeed(t *testing.T) {
	n := rand.Intn(100) + 10
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

		_, err = json.Marshal(tree)
		if err != nil {
			t.Fatalf("hash merkle tree error: %s", err)
		}
	}
}

func Test_Merkle_Unmarshal_Single_Succeed(t *testing.T) {
	path := "merkle_success.json"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s error: %s", path, err)
	}

	var newtree merkle.MerkleTree
	err = json.Unmarshal(data, &newtree)
	if err != nil {
		t.Fatalf("unmarshal merkle tree error: %s", err)
	}
	print(&newtree, t)
}

func Test_Merkle_Unmarshal_Batch_Succeed(t *testing.T) {
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

		var newtree merkle.MerkleTree
		err = json.Unmarshal(data, &newtree)
		if err != nil {
			t.Fatalf("unmarshal merkle tree error: %s", err)
		}
		// print(&newtree, t)
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

		var tree merkle.MerkleTree
		err = json.Unmarshal(data, &tree)

		rowUnmarshalError, ok := err.(merkle.RowUnmarshalError)
		if ok && rowUnmarshalError.Err != expect {
			t.Fatalf("unmarshal merkle tree, expect: %v, actual: %v", expect, rowUnmarshalError)
		}

		nodeUnmarshalError, ok := err.(merkle.NodeUnmarshalError)
		if ok && nodeUnmarshalError.Err != expect {
			t.Fatalf("unmarshal merkle tree, expect: %v, actual: %v", expect, nodeUnmarshalError)
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

	var tree merkle.MerkleTree
	err = json.Unmarshal(data, &tree)
	if err != nil {
		t.Fatalf("unmarshal merkle tree error: %s", err)
	}

	valid, err := tree.Validate()
	if err != nil {
		t.Fatalf("validate merkle tree error: %s", err)
	}
	if !valid {
		t.Fatalf("validate merkle tree should success, but failed")
	}
}

func Test_Merkle_Validate_Fail_From_Json(t *testing.T) {
	data, err := os.ReadFile("merkle_validate_failed.json")
	if err != nil {
		t.Fatalf("read json file error: %s", err)
	}

	var tree merkle.MerkleTree
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

func print(tree *merkle.MerkleTree, t *testing.T) {
	nodes := tree.Table[0]

	for len(nodes) > 1 {
		parents := make([]*merkle.MerkleTreeNode, 0, len(nodes)/2)
		for i := 0; i+1 < len(nodes); i += 2 {
			t.Logf("%x", nodes[i].Hash)
			t.Logf("%x", nodes[i+1].Hash)

			if nodes[i].Parent != nil {
				parents = append(parents, nodes[i].Parent)
			} else if nodes[i+1].Parent != nil {
				parents = append(parents, nodes[i+1].Parent)
			}
		}

		if len(nodes)%2 != 0 {
			parents = append(parents, nodes[len(nodes)-1])
		}

		nodes = parents
		t.Log("\n")
	}

	t.Logf("%x\n", nodes[0].Hash)
}
