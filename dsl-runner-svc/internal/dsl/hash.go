package dsl

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

func Sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func CanonicalJSON(data json.RawMessage) (string, error) {
	if len(data) == 0 {
		return "null", nil
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return "", err
	}
	buf := &jsonBuffer{}
	if err := encodeCanonical(buf, v); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type jsonBuffer struct {
	b []byte
}

func (b *jsonBuffer) WriteString(s string) {
	b.b = append(b.b, s...)
}

func (b *jsonBuffer) WriteByte(c byte) {
	b.b = append(b.b, c)
}

func (b *jsonBuffer) String() string {
	return string(b.b)
}

func encodeCanonical(buf *jsonBuffer, v any) error {
	switch val := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case float64:
		buf.WriteString(trimFloat(val))
	case string:
		encoded, _ := json.Marshal(val)
		buf.WriteString(string(encoded))
	case []any:
		buf.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := encodeCanonical(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case map[string]any:
		buf.WriteByte('{')
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			keyEnc, _ := json.Marshal(k)
			buf.WriteString(string(keyEnc))
			buf.WriteByte(':')
			if err := encodeCanonical(buf, val[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	default:
		encoded, _ := json.Marshal(val)
		buf.WriteString(string(encoded))
	}
	return nil
}

func trimFloat(v float64) string {
	encoded, _ := json.Marshal(v)
	return string(encoded)
}
