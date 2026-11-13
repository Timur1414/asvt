package crypto

import (
	"bytes"
	"encoding/hex"
	"errors"
)

type MerkleProof struct {
	FrameIndex int      `json:"frame_index"`
	AuditPath  []string `json:"audit_path"`
	IsLeftNode []bool   `json:"is_left_node"`
}

type MerkleTree struct {
	Leaves   [][]byte
	Layers   [][][]byte
	RootHash []byte
}

func BuildMerkleTree(frameHashes [][]byte) (*MerkleTree, error) {
	if len(frameHashes) == 0 {
		return nil, errors.New("cannot build merkle tree from empty hashes")
	}
	leaves := make([][]byte, len(frameHashes))
	copy(leaves, frameHashes)
	var layers [][][]byte
	currentLayer := leaves
	layers = append(layers, currentLayer)
	for len(currentLayer) > 1 {
		var nextLayer [][]byte
		for i := 0; i < len(currentLayer); i += 2 {
			if i+1 < len(currentLayer) {
				combined := append(currentLayer[i], currentLayer[i+1]...)
				nextLayer = append(nextLayer, Streebog256(combined))
			} else {
				combined := append(currentLayer[i], currentLayer[i]...)
				nextLayer = append(nextLayer, Streebog256(combined))
			}
		}
		layers = append(layers, nextLayer)
		currentLayer = nextLayer
	}
	return &MerkleTree{
		Leaves:   leaves,
		Layers:   layers,
		RootHash: currentLayer[0],
	}, nil
}

func (obj *MerkleTree) GenerateProof(frameIndex int) (*MerkleProof, error) {
	if frameIndex < 0 || frameIndex >= len(obj.Leaves) {
		return nil, errors.New("frame index out of bounds")
	}
	var auditPath []string
	var isLeftNode []bool
	idx := frameIndex
	for layerIdx := 0; layerIdx < len(obj.Layers)-1; layerIdx++ {
		layer := obj.Layers[layerIdx]
		var pairIdx int
		var isLeft bool
		if idx%2 == 0 {
			pairIdx = idx + 1
			if pairIdx >= len(layer) {
				pairIdx = idx
			}
			isLeft = false
		} else {
			pairIdx = idx - 1
			isLeft = true
		}
		auditPath = append(auditPath, hex.EncodeToString(layer[pairIdx]))
		isLeftNode = append(isLeftNode, isLeft)
		idx = idx / 2
	}
	return &MerkleProof{
		FrameIndex: frameIndex,
		AuditPath:  auditPath,
		IsLeftNode: isLeftNode,
	}, nil
}

func VerifyFrameProof(rootHash []byte, frameHash []byte, proof *MerkleProof) (bool, error) {
	curr := frameHash
	for i, neighborHex := range proof.AuditPath {
		neighbor, err := hex.DecodeString(neighborHex)
		if err != nil {
			return false, err
		}

		var combined []byte
		if proof.IsLeftNode[i] {
			combined = append(neighbor, curr...)
		} else {
			combined = append(curr, neighbor...)
		}
		curr = Streebog256(combined)
	}

	return bytes.Equal(curr, rootHash), nil
}
