package merkle

import (
	"Bitcoin/src/cryptography"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
)

const (
	ErrMtNodeSize      = "the node size is wrong"
	ErrMtFewRows       = "too few rows"
	ErrMtFewParentHash = "too few parent hash"
	ErrMtDuplicateHash = "duplicate Hash"
	ErrMtNoParentHash  = "no parent Hash"
)

type RowUnmarshalError struct {
	Err string
	Row int
}

func (err RowUnmarshalError) Error() string {
	return fmt.Sprintf("%v, row: %d", err.Err, err.Row)
}

type NodeUnmarshalError struct {
	Err        string
	Hash       []byte
	ParentHash []byte
}

func (err NodeUnmarshalError) Error() string {
	return fmt.Sprintf("%v, hash: %x, parent: %x", err.Err, err.Hash, err.ParentHash)
}

type MerkleTree struct {
	Table [][]*MerkleTreeNode `json:"table,omitempty"`
}

func (tree *MerkleTree) UnmarshalJSON(data []byte) error {
	var s struct {
		Table [][]*MerkleTreeNode `json:"table,omitempty"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	tree.Table = s.Table
	return rebuild(tree)
}

type MerkleTreeNode struct {
	Hash       []byte          `json:"hash,omitempty"`
	ParentHash []byte          `json:"parent,omitempty"`
	Parent     *MerkleTreeNode `json:"-"`
}

func (node *MerkleTreeNode) MarshalJSON() ([]byte, error) {
	var s = struct {
		Hash       string `json:"hash,omitempty"`
		ParentHash string `json:"parent,omitempty"`
	}{
		Hash:       hex.EncodeToString(node.Hash),
		ParentHash: hex.EncodeToString(node.ParentHash),
	}
	return json.Marshal(s)
}

func (node *MerkleTreeNode) UnmarshalJSON(data []byte) error {
	var s struct {
		Hash       string `json:"hash,omitempty"`
		ParentHash string `json:"parent,omitempty"`
	}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	node.Hash, err = hex.DecodeString(s.Hash)
	if err != nil {
		return err
	}

	node.ParentHash, err = hex.DecodeString(s.ParentHash)
	if err != nil {
		return err
	}

	return err
}

func BuildTree[T any](vals []T) (*MerkleTree, error) {
	table := make([][]*MerkleTreeNode, 0)
	nodes := make([]*MerkleTreeNode, len(vals))

	for i := 0; i < len(vals); i++ {
		hash, err := cryptography.Hash(vals[i])
		if err != nil {
			return nil, err
		}
		nodes[i] = &MerkleTreeNode{Hash: hash}
	}

	table = append(table, nodes)

	for len(nodes) > 1 {
		parents := make([]*MerkleTreeNode, len(nodes)/2+len(nodes)%2)
		row := make([]*MerkleTreeNode, len(nodes)/2)

		var i int
		for i = 0; i+1 < len(nodes); i += 2 {
			hash, err := computeParentHash(nodes[i], nodes[i+1])
			if err != nil {
				return nil, err
			}

			parent := &MerkleTreeNode{Hash: hash}
			nodes[i].ParentHash, nodes[i+1].ParentHash = hash, hash
			nodes[i].Parent, nodes[i+1].Parent = parent, parent
			parents[i/2], row[i/2] = parent, parent
		}

		table = append(table, row)

		if i < len(nodes) {
			parents[len(parents)-1] = nodes[i]
		}
		nodes = parents
	}

	tree := &MerkleTree{Table: table}

	return tree, nil
}

func (tree *MerkleTree) Validate() (bool, error) {
	return validate(tree.Table[0]...)
}

func (tree *MerkleTree) Has(hash []byte) (bool, error) {
	if len(tree.Table) == 0 {
		return false, nil
	}

	var i int
	for i = 0; i < len(tree.Table[0]); i++ {
		if bytes.Equal(tree.Table[0][i].Hash, hash) {
			break
		}
	}

	if i == len(tree.Table[0]) {
		return false, nil
	}

	nodes := make([]*MerkleTreeNode, 0, 3)

	if i%2 == 0 {
		if i+1 > len(tree.Table[0]) {
			nodes = append(nodes, tree.Table[0][i-2])
			nodes = append(nodes, tree.Table[0][i-1])
			nodes = append(nodes, tree.Table[0][i])
		} else {
			nodes = append(nodes, tree.Table[0][i])
			nodes = append(nodes, tree.Table[0][i+1])
		}
	} else {
		nodes = append(nodes, tree.Table[0][i-1])
		nodes = append(nodes, tree.Table[0][i])
	}

	return validate(nodes...)
}

func rebuild(tree *MerkleTree) error {
	if len(tree.Table) <= 1 {
		return RowUnmarshalError{Err: ErrMtFewRows}
	}

	hashmap := make(map[string]string)
	parents := make(map[string]*MerkleTreeNode)

	i, remainder := 0, 0
	for i = 1; i < len(tree.Table); i++ {
		if (len(tree.Table[i-1])+remainder)/2 != len(tree.Table[i]) {
			return RowUnmarshalError{Err: ErrMtNodeSize, Row: i}
		}

		for j := 0; j < len(tree.Table[i]); j++ {
			n := tree.Table[i][j]
			parents[string(n.Hash)] = n
		}
		remainder = (len(tree.Table[i-1]) + remainder) % 2
	}

	for i = 0; i < len(tree.Table)-1; i++ {
		parentHashs := make(map[string]string)
		for j := 0; j < len(tree.Table[i]); j++ {
			n := tree.Table[i][j]

			_, hasHash := hashmap[string(n.Hash)]
			if hasHash {
				return NodeUnmarshalError{Err: ErrMtDuplicateHash, Hash: n.Hash}
			}

			parent, hasParent := parents[string(n.ParentHash)]
			if !hasParent {
				return NodeUnmarshalError{Err: ErrMtNoParentHash, Hash: n.Hash, ParentHash: n.ParentHash}
			}

			parentHashs[string(parent.Hash)] = string(parent.Hash)

			n.Parent = parent
		}

		if len(parentHashs) < len(tree.Table[i+1]) {
			return RowUnmarshalError{Err: ErrMtFewParentHash, Row: i}
		}
	}

	n := tree.Table[i][0]
	_, hasHash := hashmap[string(n.Hash)]
	if hasHash {
		return NodeUnmarshalError{Err: ErrMtDuplicateHash, Hash: n.Hash}
	}

	return nil
}

func validate(nodes ...*MerkleTreeNode) (bool, error) {
	for len(nodes) > 1 {
		parents := make([]*MerkleTreeNode, len(nodes)/2+len(nodes)%2)

		for i := 0; i+1 < len(nodes); i += 2 {
			if !bytes.Equal(nodes[i].Parent.Hash, nodes[i+1].Parent.Hash) {
				log.Printf("parent of nodes %d and nodes %d are not same", i, i+1)
				return false, nil
			}

			hash, err := computeParentHash(nodes[i], nodes[i+1])
			if err != nil {
				return false, err
			}

			if !bytes.Equal(hash, nodes[i].Parent.Hash) {
				log.Printf("the hash of parent of nodes %d mismatch", i)
				return false, nil
			}

			parents[i/2] = nodes[i].Parent
		}

		if len(nodes)%2 != 0 {
			parents[len(parents)-1] = nodes[len(nodes)-1]
		}

		nodes = parents
	}

	return true, nil
}

func computeParentHash(nodes ...*MerkleTreeNode) ([]byte, error) {
	if len(nodes) <= 1 {
		log.Fatalf("few nodes to compute parent hash")
	}
	hashlist := make([][]byte, len(nodes))
	for i := 0; i < len(nodes); i++ {
		hashlist[i] = nodes[i].Hash
	}
	val := bytes.Join(hashlist, []byte(""))
	return cryptography.Hash(hex.EncodeToString(val))
}
