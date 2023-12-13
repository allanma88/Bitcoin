package collection

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
	ErrMtNoSiblingHash = "no sibling Hash"
)

type RowUnmarshalError struct {
	Err string
	Row int
}

func (err RowUnmarshalError) Error() string {
	return fmt.Sprintf("%v, row: %d", err.Err, err.Row)
}

type NodeUnmarshalError struct {
	Err         string
	Hash        []byte
	ParentHash  []byte
	SiblingHash []byte
}

func (err NodeUnmarshalError) Error() string {
	return fmt.Sprintf("%v, hash: %x, parent: %x, sibling: %x", err.Err, err.Hash, err.ParentHash, err.SiblingHash)
}

type MerkleTree[T any] struct {
	Table [][]*MerkleTreeNode[T] `json:"table,omitempty"`
}

func (tree *MerkleTree[T]) UnmarshalJSON(data []byte) error {
	var s struct {
		Table [][]*MerkleTreeNode[T] `json:"table,omitempty"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	tree.Table = s.Table
	return rebuild(tree)
}

type MerkleTreeNode[T any] struct {
	Hash        []byte             `json:"hash,omitempty"`
	Val         T                  `json:"val,omitempty"`
	ParentHash  []byte             `json:"parent,omitempty"`
	SiblingHash []byte             `json:"sibling,omitempty"`
	Parent      *MerkleTreeNode[T] `json:"-"`
	Sibling     *MerkleTreeNode[T] `json:"-"`
}

func (node *MerkleTreeNode[T]) MarshalJSON() ([]byte, error) {
	var s = struct {
		Hash        string `json:"hash,omitempty"`
		ParentHash  string `json:"parent,omitempty"`
		SiblingHash string `json:"sibling,omitempty"`
	}{
		Hash:        hex.EncodeToString(node.Hash),
		ParentHash:  hex.EncodeToString(node.ParentHash),
		SiblingHash: hex.EncodeToString(node.SiblingHash),
	}
	return json.Marshal(s)
}

func (node *MerkleTreeNode[T]) UnmarshalJSON(data []byte) error {
	var s struct {
		Hash        string `json:"hash,omitempty"`
		ParentHash  string `json:"parent,omitempty"`
		SiblingHash string `json:"sibling,omitempty"`
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

	node.SiblingHash, err = hex.DecodeString(s.SiblingHash)
	if err != nil {
		return err
	}

	return err
}

func (node *MerkleTreeNode[T]) setParent(parent *MerkleTreeNode[T]) {
	node.ParentHash = parent.Hash
	node.Parent = parent
}

func (node *MerkleTreeNode[T]) setSibling(sibling *MerkleTreeNode[T]) {
	node.SiblingHash = sibling.Hash
	node.Sibling = sibling
}

func (node *MerkleTreeNode[T]) validate() (bool, error) {
	for node.Parent != nil {
		if node.Sibling == nil {
			// log.Printf("%x: no sibling", node.Hash)
			return false, nil
		}

		if !bytes.Equal(node.ParentHash, node.Sibling.ParentHash) {
			// log.Printf("%x: parent hash %x mismatch with sibling %x", node.Hash, node.ParentHash, node.Sibling.ParentHash)
			return false, nil
		}

		parentHash, err := node.computeParentHash(node.Sibling)
		if err != nil {
			return false, err
		}

		parentHash1, err := node.Sibling.computeParentHash(node)
		if err != nil {
			return false, err
		}

		if !bytes.Equal(node.ParentHash, parentHash) && !bytes.Equal(node.ParentHash, parentHash1) {
			// log.Printf("%x: compute parent hash %x mismatch with the parent hash %x or %x", node.Hash, node.ParentHash, parentHash, parentHash1)
			return false, nil
		}

		node = node.Parent
	}

	return true, nil
}

func (node *MerkleTreeNode[T]) computeParentHash(sibling *MerkleTreeNode[T]) ([]byte, error) {
	hashlist := make([][]byte, 2)
	hashlist[0] = node.Hash
	hashlist[1] = sibling.Hash

	val := bytes.Join(hashlist, []byte(""))
	return cryptography.Hash(hex.EncodeToString(val))
}

func BuildTree[T any](vals []T) (*MerkleTree[T], error) {
	table := make([][]*MerkleTreeNode[T], 0)
	nodes := make([]*MerkleTreeNode[T], len(vals))

	for i := 0; i < len(vals); i++ {
		hash, err := cryptography.Hash(vals[i])
		if err != nil {
			return nil, err
		}
		nodes[i] = &MerkleTreeNode[T]{
			Hash: hash,
			Val:  vals[i],
		}
	}

	table = append(table, nodes)

	for len(nodes) > 1 {
		parents := make([]*MerkleTreeNode[T], len(nodes)/2+len(nodes)%2)
		row := make([]*MerkleTreeNode[T], len(nodes)/2)

		var i int
		for i = 0; i+1 < len(nodes); i += 2 {
			hash, err := nodes[i].computeParentHash(nodes[i+1])
			if err != nil {
				return nil, err
			}

			parent := &MerkleTreeNode[T]{Hash: hash}
			nodes[i].setParent(parent)
			nodes[i+1].setParent(parent)
			nodes[i].setSibling(nodes[i+1])
			nodes[i+1].setSibling(nodes[i])
			parents[i/2], row[i/2] = parent, parent
		}

		table = append(table, row)

		if i < len(nodes) {
			parents[len(parents)-1] = nodes[i]
		}
		nodes = parents
	}

	tree := &MerkleTree[T]{Table: table}

	return tree, nil
}

func (tree *MerkleTree[T]) Validate() (bool, error) {
	nodes := tree.Table[0]

	for len(nodes) > 1 {
		parents := make([]*MerkleTreeNode[T], len(nodes)/2+len(nodes)%2)

		for i := 0; i+1 < len(nodes); i += 2 {
			if !bytes.Equal(nodes[i].Parent.Hash, nodes[i+1].Parent.Hash) {
				log.Printf("parent of nodes %d and nodes %d are not same, total %d", i, i+1, len(nodes))
				return false, nil
			}

			hash, err := nodes[i].computeParentHash(nodes[i+1])
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

func (tree *MerkleTree[T]) Get(hash []byte) (T, error) {
	if len(tree.Table) == 0 {
		return *new(T), nil
	}

	var i int
	for i = 0; i < len(tree.Table[0]); i++ {
		if bytes.Equal(tree.Table[0][i].Hash, hash) {
			break
		}
	}

	if i == len(tree.Table[0]) {
		return *new(T), nil
	}

	node := tree.Table[0][i]
	valid, err := node.validate()
	if err != nil {
		return *new(T), err
	}
	if valid {
		return node.Val, nil
	} else {
		return *new(T), nil
	}
}

func (tree *MerkleTree[T]) GetVals() []T {
	if len(tree.Table) == 0 {
		return nil
	}

	vals := make([]T, len(tree.Table[0]))
	for i := 0; i < len(tree.Table[0]); i++ {
		vals[i] = tree.Table[0][i].Val
	}
	return vals
}

func rebuild[T any](tree *MerkleTree[T]) error {
	if len(tree.Table) <= 1 {
		return RowUnmarshalError{Err: ErrMtFewRows}
	}

	nodes := make(map[string]*MerkleTreeNode[T])
	for i := 0; i < len(tree.Table); i++ {
		for j := 0; j < len(tree.Table[i]); j++ {
			n := tree.Table[i][j]
			_, hasHash := nodes[string(n.Hash)]
			if hasHash {
				return NodeUnmarshalError{Err: ErrMtDuplicateHash, Hash: n.Hash}
			}

			nodes[string(n.Hash)] = n
		}
	}

	i, remainder := 0, 0
	for i = 0; i < len(tree.Table)-1; i++ {
		if (len(tree.Table[i])+remainder)/2 != len(tree.Table[i+1]) {
			return RowUnmarshalError{Err: ErrMtNodeSize, Row: i + 1}
		}

		parentHashs := make(map[string]string)
		for j := 0; j < len(tree.Table[i]); j++ {
			n := tree.Table[i][j]

			parent, hasParent := nodes[string(n.ParentHash)]
			if !hasParent {
				return NodeUnmarshalError{Err: ErrMtNoParentHash, Hash: n.Hash, ParentHash: n.ParentHash}
			}
			n.Parent = parent

			parentHashs[string(parent.Hash)] = string(parent.Hash)

			sibling, hasSibling := nodes[string(n.SiblingHash)]
			if !hasSibling {
				return NodeUnmarshalError{Err: ErrMtNoSiblingHash, Hash: n.Hash, SiblingHash: n.SiblingHash}
			}
			n.Sibling = sibling
		}

		if len(parentHashs) < len(tree.Table[i+1]) {
			return RowUnmarshalError{Err: ErrMtFewParentHash, Row: i}
		}

		remainder = (len(tree.Table[i]) + remainder) % 2
	}

	return nil
}
