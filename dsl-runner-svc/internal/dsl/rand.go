package dsl

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type Rand struct {
	state uint64
}

func NewRand(seed string) *Rand {
	sum := sha256.Sum256([]byte(seed))
	state := binary.LittleEndian.Uint64(sum[:8])
	if state == 0 {
		state = 0x853c49e6748fea9b
	}
	return &Rand{state: state}
}

func (r *Rand) next() uint64 {
	// xorshift64*
	x := r.state
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	r.state = x
	return x * 2685821657736338717
}

func (r *Rand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next() % uint64(n))
}

func (r *Rand) Float64() float64 {
	return float64(r.next()>>11) * (1.0 / (1 << 53))
}

func (r *Rand) String(alpha string, length int) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("invalid length")
	}
	alphabet := alphabetFor(alpha)
	if alphabet == "" {
		return "", fmt.Errorf("unknown alphabet")
	}
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(buf), nil
}

func alphabetFor(name string) string {
	switch name {
	case "alnum":
		return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	case "hex":
		return "0123456789abcdef"
	case "alpha":
		return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	default:
		return ""
	}
}

func (r *Rand) UUID() string {
	b := make([]byte, 16)
	for i := 0; i < len(b); i += 8 {
		val := r.next()
		binary.LittleEndian.PutUint64(b[i:], val)
	}
	// Set version 4 and variant 10xx
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return formatUUID(b)
}

func formatUUID(b []byte) string {
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])
	return string(buf)
}
