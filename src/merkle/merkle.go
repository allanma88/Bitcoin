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
	Hash        []byte          `json:"hash,omitempty"`
	ParentHash  []byte          `json:"parent,omitempty"`
	SiblingHash []byte          `json:"sibling,omitempty"`
	Parent      *MerkleTreeNode `json:"-"`
	Sibling     *MerkleTreeNode `json:"-"`
}

func (node *MerkleTreeNode) MarshalJSON() ([]byte, error) {
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

func (node *MerkleTreeNode) UnmarshalJSON(data []byte) error {
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

func (node *MerkleTreeNode) SetParent(parent *MerkleTreeNode) {
	node.ParentHash = parent.Hash
	node.Parent = parent
}

func (node *MerkleTreeNode) SetSibling(sibling *MerkleTreeNode) {
	node.SiblingHash = sibling.Hash
	node.Sibling = sibling
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
			nodes[i].SetParent(parent)
			nodes[i+1].SetParent(parent)
			nodes[i].SetSibling(nodes[i+1])
			nodes[i+1].SetSibling(nodes[i])
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
	nodes := tree.Table[0]

	for len(nodes) > 1 {
		parents := make([]*MerkleTreeNode, len(nodes)/2+len(nodes)%2)

		for i := 0; i+1 < len(nodes); i += 2 {
			if !bytes.Equal(nodes[i].Parent.Hash, nodes[i+1].Parent.Hash) {
				log.Printf("parent of nodes %d and nodes %d are not same, total %d", i, i+1, len(nodes))
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

	return tree.Table[0][i].validate()
}

func rebuild(tree *MerkleTree) error {
	if len(tree.Table) <= 1 {
		return RowUnmarshalError{Err: ErrMtFewRows}
	}

	nodes := make(map[string]*MerkleTreeNode)
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

func (node *MerkleTreeNode) validate() (bool, error) {
	for node.Parent != nil {
		if node.Sibling == nil {
			// log.Printf("%x: no sibling", node.Hash)
			return false, nil
		}

		if !bytes.Equal(node.ParentHash, node.Sibling.ParentHash) {
			// log.Printf("%x: parent hash %x mismatch with sibling %x", node.Hash, node.ParentHash, node.Sibling.ParentHash)
			return false, nil
		}

		parentHash, err := computeParentHash(node, node.Sibling)
		if err != nil {
			return false, err
		}

		parentHash1, err := computeParentHash(node.Sibling, node)
		if err != nil {
			return false, err
		}

		if !bytes.Equal(node.ParentHash, parentHash) && !bytes.Equal(node.ParentHash, parentHash1) {
			log.Printf("%x: compute parent hash %x mismatch with the parent hash %x or %x", node.Hash, node.ParentHash, parentHash, parentHash1)
			return false, err
		}

		node = node.Parent
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
