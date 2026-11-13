package crypto

import (
	"encoding/binary"
	"errors"
	"hash"
)

const (
	BlockSize = 64
	Size256   = 32
	Size512   = 64
)

// Таблица перестановок ГОСТ Р 34.11-2012
var piTable = [256]byte{
	0xFC, 0xEE, 0xDD, 0x11, 0xCF, 0x6E, 0x31, 0x16, 0xFB, 0xC4, 0xFA, 0xDA, 0x23, 0xC5, 0x04, 0x4D,
	0xE9, 0x77, 0xF0, 0xDB, 0x93, 0x2E, 0x99, 0xBA, 0x17, 0x36, 0xF1, 0xBB, 0x14, 0xCD, 0x5F, 0xC1,
	0xF9, 0x18, 0x65, 0x5A, 0xE2, 0x5C, 0xEF, 0x21, 0x81, 0x1C, 0x3C, 0x42, 0x8B, 0x01, 0x8E, 0x4F,
	0x05, 0x84, 0x02, 0xAE, 0xE3, 0x67, 0x8F, 0x51, 0xEB, 0x63, 0xF4, 0x96, 0xBF, 0x03, 0xE0, 0x52,
	0x3B, 0xD6, 0xD2, 0xAE, 0xE7, 0x6B, 0x33, 0xF7, 0x39, 0x45, 0x40, 0xC3, 0x30, 0x07, 0x35, 0x60,
	0x51, 0x83, 0x38, 0x82, 0x53, 0x71, 0xB8, 0x26, 0x47, 0xCE, 0xE7, 0x72, 0xA0, 0xE6, 0x71, 0xB0,
	0x54, 0x50, 0x85, 0x23, 0x84, 0x18, 0x08, 0x4C, 0x18, 0x80, 0xA7, 0x57, 0x85, 0x13, 0x44, 0x77,
	0xC2, 0xB2, 0x2F, 0x80, 0x73, 0x55, 0x97, 0x31, 0x11, 0x14, 0x1F, 0xF8, 0x3A, 0x2B, 0xC2, 0x00,
	0x01, 0x08, 0x0A, 0x27, 0x66, 0x09, 0x49, 0xF2, 0x26, 0x48, 0x33, 0xB0, 0xE2, 0xD5, 0xD9, 0x28,
	0x7F, 0x09, 0x76, 0x24, 0xA6, 0x32, 0x72, 0x1E, 0x6C, 0x62, 0x9D, 0x85, 0x10, 0x1A, 0xB8, 0x38,
	0x39, 0x57, 0x8E, 0x98, 0x54, 0xDB, 0x85, 0xA1, 0xEA, 0xD8, 0x42, 0x94, 0x99, 0x0D, 0x38, 0x60,
	0x73, 0x26, 0x01, 0x0B, 0x20, 0x01, 0x1F, 0xF4, 0x42, 0x04, 0x53, 0xD5, 0x63, 0x5A, 0x14, 0xF3,
	0x67, 0xA2, 0x5D, 0x80, 0xCA, 0x71, 0x86, 0x08, 0x9A, 0xAE, 0x40, 0xA2, 0x08, 0x0C, 0x17, 0x29,
	0x1D, 0x3E, 0x6A, 0x4B, 0x7F, 0xAB, 0xE6, 0x79, 0x37, 0xC8, 0x2B, 0x6E, 0xD6, 0x8B, 0x35, 0x64,
	0x14, 0x53, 0x56, 0xA0, 0xD2, 0x8A, 0x3B, 0xB1, 0x2D, 0x18, 0x26, 0x41, 0xF6, 0x46, 0x97, 0x6F,
	0xB4, 0x2E, 0x00, 0x0F, 0xB8, 0x09, 0x94, 0x52, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45,
}

var lMatrix = [8]uint64{
	0x011860830a411f1f,
	0x1f011860830a411f,
	0x1f1f011860830a41,
	0x411f1f011860830a,
	0x0a411f1f01186083,
	0x830a411f1f011860,
	0x60830a411f1f0118,
	0x1860830a411f1f01,
}

var precomputedTable [8][256]uint64

func init() {
	for i := 0; i < 8; i++ {
		for j := 0; j < 256; j++ {
			b := piTable[j]
			var v uint64
			for k := 0; k < 8; k++ {
				if (b>>k)&1 != 0 {
					v ^= lMatrix[k]
				}
			}
			precomputedTable[i][j] = (v << (8 * i)) | (v >> (64 - 8*i))
		}
	}
}

type digest struct {
	h      [8]uint64
	n      [8]uint64
	sigma  [8]uint64
	buf    [BlockSize]byte
	bufLen int
	is256  bool
}

// New256 возвращает интерфейс hash.Hash для Стрибог-256
func New256() hash.Hash {
	d := &digest{is256: true}
	d.Reset()
	return d
}

func (d *digest) BlockSize() int { return BlockSize }

func (d *digest) Size() int {
	if d.is256 {
		return Size256
	}
	return Size512
}

func (d *digest) Reset() {
	var ivVal uint64
	if d.is256 {
		ivVal = 0x0101010101010101
	} else {
		ivVal = 0x0000000000000000
	}
	for i := range d.h {
		d.h[i] = ivVal
		d.n[i] = 0
		d.sigma[i] = 0
	}
	d.bufLen = 0
}

func (d *digest) Write(p []byte) (nn int, err error) {
	nn = len(p)
	if d.bufLen > 0 {
		n := copy(d.buf[d.bufLen:], p)
		d.bufLen += n
		p = p[n:]
		if d.bufLen == BlockSize {
			d.processBlock(d.buf[:])
			d.bufLen = 0
		}
	}
	for len(p) >= BlockSize {
		d.processBlock(p[:BlockSize])
		p = p[BlockSize:]
	}
	if len(p) > 0 {
		d.bufLen = copy(d.buf[:], p)
	}
	return nn, nil
}

func (d *digest) processBlock(m []byte) {
	var mVec [8]uint64
	for i := 0; i < 8; i++ {
		mVec[i] = binary.LittleEndian.Uint64(m[i*8:])
	}
	d.gN(&d.h, &mVec, &d.n)
	d.addN(512)
	d.addSigma(&mVec)
}

func (d *digest) addN(bits uint64) {
	carry := bits
	for i := 0; i < 8 && carry > 0; i++ {
		sum := d.n[i] + carry
		if sum < d.n[i] {
			carry = 1
		} else {
			carry = 0
		}
		d.n[i] = sum
	}
}

func (d *digest) addSigma(m *[8]uint64) {
	var carry uint64
	for i := 0; i < 8; i++ {
		sum := d.sigma[i] + m[i] + carry
		if sum < d.sigma[i] || (carry > 0 && sum == d.sigma[i]) {
			carry = 1
		} else {
			carry = 0
		}
		d.sigma[i] = sum
	}
}

func (d *digest) gN(h, m, n *[8]uint64) {
	var k, state [8]uint64
	for i := 0; i < 8; i++ {
		k[i] = h[i] ^ n[i]
		state[i] = m[i] ^ k[i]
	}
	for round := 0; round < 12; round++ {
		state = transformLPS(&state)
		k = transformLPS(&k)
		k[0] ^= roundConstants[round]
		for i := 0; i < 8; i++ {
			state[i] ^= k[i]
		}
	}
	for i := 0; i < 8; i++ {
		h[i] ^= state[i] ^ m[i]
	}
}

func transformLPS(state *[8]uint64) [8]uint64 {
	var res [8]uint64
	for i := 0; i < 8; i++ {
		v := state[i]
		for j := 0; j < 8; j++ {
			b := byte(v >> (j * 8))
			res[j] ^= precomputedTable[j][b]
		}
	}
	return res
}

var roundConstants = [12]uint64{
	0x0100000000000000, 0x0200000000000000, 0x0300000000000000, 0x0400000000000000,
	0x0500000000000000, 0x0600000000000000, 0x0700000000000000, 0x0800000000000000,
	0x0900000000000000, 0x0a00000000000000, 0x0b00000000000000, 0x0c00000000000000,
}

func (d *digest) Sum(b []byte) []byte {
	clone := *d
	hashVal := clone.checkSum()
	if d.is256 {
		return append(b, hashVal[32:]...)
	}
	return append(b, hashVal[:]...)
}

func (d *digest) checkSum() [64]byte {
	var pad [BlockSize]byte
	copy(pad[:], d.buf[:d.bufLen])
	pad[d.bufLen] = 0x01
	var mVec [8]uint64
	for i := 0; i < 8; i++ {
		mVec[i] = binary.LittleEndian.Uint64(pad[i*8:])
	}
	d.gN(&d.h, &mVec, &d.n)
	d.addN(uint64(d.bufLen * 8))
	d.addSigma(&mVec)
	var zero [8]uint64
	d.gN(&d.h, &d.n, &zero)
	d.gN(&d.h, &d.sigma, &zero)
	var out [64]byte
	for i := 0; i < 8; i++ {
		binary.LittleEndian.PutUint64(out[i*8:], d.h[i])
	}
	return out
}

func Streebog256(data []byte) []byte {
	h := New256()
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func HashFrame(frameData []byte) ([]byte, error) {
	if len(frameData) == 0 {
		return nil, errors.New("empty frame data")
	}
	return Streebog256(frameData), nil
}
