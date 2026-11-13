package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

// SegmentManifest паспорт отдельного фрагмента видеопотока
type SegmentManifest struct {
	SegmentName    string   `json:"segment_name"`    // Имя файла (например, segment_00001.ts)
	SequenceNumber uint64   `json:"seq_num"`         // Порядковый номер
	TimestampUTC   string   `json:"timestamp_utc"`   // Метка времени начала фрагмента (ISO 8601)
	DurationSec    float64  `json:"duration_sec"`    // Длительность фрагмента в секундах
	PrevBlockHash  string   `json:"prev_block_hash"` // Хэш предыдущего фрагмента H_(i-1)
	MerkleRoot     string   `json:"merkle_root"`     // Корень дерева кадров
	FrameCount     int      `json:"frame_count"`     // Количество кадров
	FrameHashes    []string `json:"frame_hashes"`    // Хэши всех кадров сегмента
	Signature      string   `json:"signature"`       // Электронная подпись метаданных (hex)
	PublicKey      string   `json:"public_key"`      // Публичный ключ устройства/камеры (hex)
}

type Signer interface {
	Sign(data []byte) ([]byte, error)
	PublicKeyBytes() []byte
}

// DefaultSigner формирует эллиптическую подпись в связке с ГОСТ Стрибог-256
type DefaultSigner struct {
	privKey *ecdsa.PrivateKey
}

func NewDefaultSigner() (*DefaultSigner, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return &DefaultSigner{privKey: priv}, nil
}

func (obj *DefaultSigner) Sign(data []byte) ([]byte, error) {
	hash := Streebog256(data)
	r, sigS, err := ecdsa.Sign(rand.Reader, obj.privKey, hash)
	if err != nil {
		return nil, err
	}
	rBytes := r.Bytes()
	sBytes := sigS.Bytes()
	sig := make([]byte, 64)
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)
	return sig, nil
}

func (obj *DefaultSigner) PublicKeyBytes() []byte {
	return elliptic.Marshal(obj.privKey.Curve, obj.privKey.PublicKey.X, obj.privKey.PublicKey.Y)
}

func VerifySignature(publicKeyBytes []byte, data []byte, sigBytes []byte) (bool, error) {
	if len(sigBytes) != 64 {
		return false, errors.New("invalid signature length, expected 64 bytes")
	}
	curve := elliptic.P256()
	x, y := elliptic.Unmarshal(curve, publicKeyBytes)
	if x == nil || y == nil {
		return false, errors.New("failed to unmarshal public key")
	}
	pubKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])
	hash := Streebog256(data)
	return ecdsa.Verify(pubKey, hash, r, s), nil
}

func BuildDataToSign(seq uint64, ts string, prevHash string, merkleRoot string) []byte {
	return []byte(fmt.Sprintf("%d|%s|%s|%s", seq, ts, prevHash, merkleRoot))
}
