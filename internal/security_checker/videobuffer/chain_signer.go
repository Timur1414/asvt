package videobuffer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Timur1414/asvt/internal/security_checker/videobuffer/crypto"
)

type ChainSigner struct {
	mu            sync.Mutex
	signer        crypto.Signer
	prevBlockHash string
	sequenceNum   uint64
}

func NewChainSigner(signer crypto.Signer) *ChainSigner {
	return &ChainSigner{
		signer:        signer,
		prevBlockHash: "GENESIS_BLOCK",
		sequenceNum:   1,
	}
}

func (obj *ChainSigner) SignTSFile(tsPath string, timestamp time.Time, durationSec float64) (*crypto.SegmentManifest, error) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	rawFrames, err := ExtractFramesFromTS(tsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract frames from %s: %w", tsPath, err)
	}
	frameHashes := make([][]byte, len(rawFrames))
	frameHashesHex := make([]string, len(rawFrames))
	for i, frameBytes := range rawFrames {
		h := crypto.Streebog256(frameBytes)
		frameHashes[i] = h
		frameHashesHex[i] = hex.EncodeToString(h)
	}
	tree, err := crypto.BuildMerkleTree(frameHashes)
	if err != nil {
		return nil, fmt.Errorf("failed to build Merkle tree: %w", err)
	}
	merkleRootHex := hex.EncodeToString(tree.RootHash)
	tsStr := timestamp.UTC().Format(time.RFC3339Nano)

	dataToSign := crypto.BuildDataToSign(obj.sequenceNum, tsStr, obj.prevBlockHash, merkleRootHex)
	sigBytes, err := obj.signer.Sign(dataToSign)
	if err != nil {
		return nil, fmt.Errorf("failed to sign segment data: %w", err)
	}
	sigHex := hex.EncodeToString(sigBytes)
	manifest := &crypto.SegmentManifest{
		SegmentName:    filepath.Base(tsPath),
		SequenceNumber: obj.sequenceNum,
		TimestampUTC:   tsStr,
		DurationSec:    durationSec,
		PrevBlockHash:  obj.prevBlockHash,
		MerkleRoot:     merkleRootHex,
		FrameCount:     len(frameHashes),
		FrameHashes:    frameHashesHex,
		Signature:      sigHex,
		PublicKey:      hex.EncodeToString(obj.signer.PublicKeyBytes()),
	}

	sigFilePath := strings.TrimSuffix(tsPath, filepath.Ext(tsPath)) + ".sig.json"
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}
	if err := os.WriteFile(sigFilePath, manifestBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write sig file %s: %w", sigFilePath, err)
	}
	obj.prevBlockHash = sigHex
	obj.sequenceNum++
	return manifest, nil
}
