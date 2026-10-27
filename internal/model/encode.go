package model

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"net/url"
	"strings"
)

// Encoding represents a supported encoding scheme.
type Encoding string

const (
	EncodingBase64    Encoding = "base64"
	EncodingBase64URL Encoding = "base64url"
	EncodingBase32    Encoding = "base32"
	EncodingBase32Hex Encoding = "base32hex"
	EncodingBase58    Encoding = "base58"
	EncodingBase85    Encoding = "base85"
	EncodingHex       Encoding = "hex"
	EncodingURLEnc    Encoding = "url"
)

// AllEncodings returns the list of supported encodings.
func AllEncodings() []Encoding {
	return []Encoding{
		EncodingBase64,
		EncodingBase64URL,
		EncodingBase32,
		EncodingBase32Hex,
		EncodingBase58,
		EncodingBase85,
		EncodingHex,
		EncodingURLEnc,
	}
}

// Encode encodes raw bytes using the specified encoding.
func Encode(data []byte, enc Encoding) (string, error) {
	switch enc {
	case EncodingBase64:
		return base64.StdEncoding.EncodeToString(data), nil
	case EncodingBase64URL:
		return base64.URLEncoding.EncodeToString(data), nil
	case EncodingBase32:
		return base32.StdEncoding.EncodeToString(data), nil
	case EncodingBase32Hex:
		return base32.HexEncoding.EncodeToString(data), nil
	case EncodingBase58:
		return Base58Encode(data), nil
	case EncodingBase85:
		return Base85Encode(data), nil
	case EncodingHex:
		return hex.EncodeToString(data), nil
	case EncodingURLEnc:
		return url.QueryEscape(string(data)), nil
	default:
		return "", ErrUnsupportedEncoding
	}
}

// Decode decodes a string using the specified encoding.
func Decode(s string, enc Encoding) ([]byte, error) {
	switch enc {
	case EncodingBase64:
		return base64.StdEncoding.DecodeString(s)
	case EncodingBase64URL:
		return base64.URLEncoding.DecodeString(s)
	case EncodingBase32:
		return base32.StdEncoding.DecodeString(s)
	case EncodingBase32Hex:
		return base32.HexEncoding.DecodeString(s)
	case EncodingBase58:
		return Base58Decode(s)
	case EncodingBase85:
		return Base85Decode(s)
	case EncodingHex:
		return hex.DecodeString(s)
	case EncodingURLEnc:
		s, err := url.QueryUnescape(s)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	default:
		return nil, ErrUnsupportedEncoding
	}
}

// Convert converts data from one encoding to another.
func Convert(s string, from, to Encoding) (string, error) {
	data, err := Decode(s, from)
	if err != nil {
		return "", err
	}
	return Encode(data, to)
}

// Detect attempts to detect the encoding of a string.
func Detect(s string) []Encoding {
	var candidates []Encoding

	// Try base64
	if isBase64(s) {
		candidates = append(candidates, EncodingBase64)
	}
	if isBase64URL(s) {
		candidates = append(candidates, EncodingBase64URL)
	}
	// Try base32
	if isBase32(s) {
		candidates = append(candidates, EncodingBase32)
	}
	// Try hex
	if isHex(s) {
		candidates = append(candidates, EncodingHex)
	}
	// Try base58
	if isBase58(s) {
		candidates = append(candidates, EncodingBase58)
	}
	// Try base85
	if isBase85(s) {
		candidates = append(candidates, EncodingBase85)
	}
	// URL encoding
	if strings.Contains(s, "%") {
		if _, err := url.QueryUnescape(s); err == nil {
			candidates = append(candidates, EncodingURLEnc)
		}
	}

	return candidates
}

// Validation helpers

func isBase64(s string) bool {
	if len(s) == 0 || len(s)%4 != 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}

func isBase64URL(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func isBase32(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7') || c == '=') {
			return false
		}
	}
	return true
}

func isHex(s string) bool {
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func isBase58(s string) bool {
	if len(s) == 0 {
		return false
	}
	const base58Chars = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	for _, c := range s {
		if !strings.ContainsRune(base58Chars, c) {
			return false
		}
	}
	return true
}

func isBase85(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '!' || c > 'u' {
			return false
		}
	}
	return true
}

// Base58Encode encodes bytes using Bitcoin Base58 alphabet.
func Base58Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	// Count leading zeros
	zeros := 0
	for zeros < len(data) && data[zeros] == 0 {
		zeros++
	}

	// Convert to big integer
	num := new(big.Int).SetBytes(data)

	// Encode
	var encoded []byte
	base := big.NewInt(58)
	mod := new(big.Int)
	for num.Sign() > 0 {
		num.DivMod(num, base, mod)
		encoded = append([]byte{alphabet[mod.Int64()]}, encoded...)
	}

	// Add leading '1' for each leading zero byte
	for i := 0; i < zeros; i++ {
		encoded = append([]byte{'1'}, encoded...)
	}

	return string(encoded)
}

// Base58Decode decodes a Base58 string to bytes.
func Base58Decode(s string) ([]byte, error) {
	if len(s) == 0 {
		return []byte{}, nil
	}
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	// Count leading '1's (leading zeros)
	zeros := 0
	for zeros < len(s) && s[zeros] == '1' {
		zeros++
	}

	// Convert to big integer
	num := new(big.Int)
	base := big.NewInt(58)
	for i := zeros; i < len(s); i++ {
		idx := strings.IndexByte(alphabet, s[i])
		if idx < 0 {
			return nil, ErrInvalidBase58
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(int64(idx)))
	}

	// Convert to bytes
	result := num.Bytes()

	// Add leading zero bytes
	if zeros > 0 {
		padded := make([]byte, zeros+len(result))
		copy(padded[zeros:], result)
		return padded, nil
	}

	return result, nil
}

// Base85Encode encodes bytes using ASCII85 (Base85).
func Base85Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var result strings.Builder
	i := 0
	for i+4 <= len(data) {
		val := uint32(data[i])<<24 | uint32(data[i+1])<<16 | uint32(data[i+2])<<8 | uint32(data[i+3])
		var buf [5]byte
		for j := 4; j >= 0; j-- {
			buf[j] = byte('!' + (val % 85))
			val /= 85
		}
		result.Write(buf[:])
		i += 4
	}

	// Handle remaining bytes
	remaining := len(data) - i
	if remaining > 0 {
		var pad [4]byte
		copy(pad[:], data[i:])
		val := uint32(pad[0])<<24 | uint32(pad[1])<<16 | uint32(pad[2])<<8 | uint32(pad[3])
		var buf [5]byte
		for j := 4; j >= 0; j-- {
			buf[j] = byte('!' + (val % 85))
			val /= 85
		}
		result.Write(buf[:remaining+1])
	}

	return result.String()
}

// Base85Decode decodes an ASCII85 (Base85) string to bytes.
func Base85Decode(s string) ([]byte, error) {
	if len(s) == 0 {
		return []byte{}, nil
	}

	var result []byte
	i := 0
	for i+5 <= len(s) {
		val := uint32(0)
		for j := 0; j < 5; j++ {
			c := s[i+j]
			if c < '!' || c > 'u' {
				return nil, ErrInvalidBase85
			}
			val = val*85 + uint32(c-'!')
		}
		result = append(result, byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
		i += 5
	}

	// Handle remaining characters
	remaining := len(s) - i
	if remaining > 0 {
		val := uint32(0)
		for j := 0; j < remaining; j++ {
			c := s[i+j]
			if c < '!' || c > 'u' {
				return nil, ErrInvalidBase85
			}
			val = val*85 + uint32(c-'!')
		}
		// Pad with 'u' (84) for missing characters
		for j := remaining; j < 5; j++ {
			val = val*85 + 84
		}
		// Only output remaining-1 bytes
		buf := []byte{byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val)}
		result = append(result, buf[:remaining-1]...)
	}

	return result, nil
}
