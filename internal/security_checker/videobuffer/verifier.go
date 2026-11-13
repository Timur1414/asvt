package videobuffer

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Timur1414/asvt/internal/security_checker/videobuffer/crypto"
)

// VideoClipPassport итоговый криптографический паспорт доказательной базы для инцидента
type VideoClipPassport struct {
	ClipID            string                   `json:"clip_id"`
	StartTime         string                   `json:"start_time"`
	EndTime           string                   `json:"end_time"`
	SegmentCount      int                      `json:"segment_count"`
	Segments          []crypto.SegmentManifest `json:"segments"`
	IntegrityVerified bool                     `json:"integrity_verified"`
}

func VerifySegmentManifest(manifest *crypto.SegmentManifest) (bool, error) {
	pubKeyBytes, err := hex.DecodeString(manifest.PublicKey)
	if err != nil {
		return false, fmt.Errorf("invalid public key hex: %w", err)
	}
	sigBytes, err := hex.DecodeString(manifest.Signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature hex: %w", err)
	}
	dataToSign := crypto.BuildDataToSign(
		manifest.SequenceNumber,
		manifest.TimestampUTC,
		manifest.PrevBlockHash,
		manifest.MerkleRoot,
	)
	return crypto.VerifySignature(pubKeyBytes, dataToSign, sigBytes)
}

func VerifyChainIntegrity(segments []crypto.SegmentManifest) (bool, error) {
	if len(segments) == 0 {
		return false, errors.New("empty segment list")
	}
	for i := 0; i < len(segments); i++ {
		valid, err := VerifySegmentManifest(&segments[i])
		if err != nil || !valid {
			return false, fmt.Errorf("signature verification failed for segment %s: %w", segments[i].SegmentName, err)
		}
		if i > 0 {
			prevSig := segments[i-1].Signature
			if segments[i].PrevBlockHash != prevSig {
				return false, fmt.Errorf("chain breakage detected between %s and %s (expected prev: %s, got: %s)",
					segments[i-1].SegmentName, segments[i].SegmentName, prevSig, segments[i].PrevBlockHash)
			}
		}
	}
	return true, nil
}

func VerifyPassportFile(passportPath string) (*VideoClipPassport, error) {
	data, err := os.ReadFile(passportPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read passport file: %w", err)
	}
	var passport VideoClipPassport
	if err := json.Unmarshal(data, &passport); err != nil {
		return nil, fmt.Errorf("failed to parse passport JSON: %w", err)
	}
	valid, err := VerifyChainIntegrity(passport.Segments)
	if err != nil || !valid {
		return &passport, fmt.Errorf("passport verification failed: %w", err)
	}
	passport.IntegrityVerified = true
	return &passport, nil
}
