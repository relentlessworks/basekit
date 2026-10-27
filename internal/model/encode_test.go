package model

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name string
		enc  Encoding
		data []byte
	}{
		{"base64-hello", EncodingBase64, []byte("hello")},
		{"base64url-hello", EncodingBase64URL, []byte("hello")},
		{"base32-hello", EncodingBase32, []byte("hello")},
		{"base32hex-hello", EncodingBase32Hex, []byte("hello")},
		{"base58-hello", EncodingBase58, []byte("hello")},
		{"base85-hello", EncodingBase85, []byte("hello")},
		{"hex-hello", EncodingHex, []byte("hello")},
		{"url-hello", EncodingURLEnc, []byte("hello world")},
		{"base64-empty", EncodingBase64, []byte{}},
		{"base64-binary", EncodingBase64, []byte{0x00, 0x01, 0xFF, 0xFE}},
		{"base58-binary", EncodingBase58, []byte{0x00, 0x00, 0x01, 0x02}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode(tt.data, tt.enc)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			decoded, err := Decode(encoded, tt.enc)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if !bytes.Equal(tt.data, decoded) {
				t.Errorf("roundtrip mismatch: got %v, want %v", decoded, tt.data)
			}
		})
	}
}

func TestBase58Roundtrip(t *testing.T) {
	tests := [][]byte{
		[]byte("hello"),
		[]byte("The quick brown fox"),
		{0x00, 0x01, 0x02, 0x03},
		{0x00, 0x00, 0x00, 0x01},
		{0xFF, 0xEE, 0xDD, 0xCC},
		[]byte(""),
	}

	for _, data := range tests {
		encoded := Base58Encode(data)
		decoded, err := Base58Decode(encoded)
		if err != nil {
			t.Errorf("Base58Decode(%q) error: %v", encoded, err)
			continue
		}
		if !bytes.Equal(data, decoded) {
			t.Errorf("Base58 roundtrip: got %v, want %v", decoded, data)
		}
	}
}

func TestBase85Roundtrip(t *testing.T) {
	tests := [][]byte{
		[]byte("hello"),
		[]byte("The quick brown fox jumps"),
		{0x00, 0x01, 0x02, 0x03},
		{0xFF, 0xEE, 0xDD, 0xCC},
		[]byte("a"),
		[]byte("ab"),
		[]byte("abc"),
		[]byte(""),
	}

	for _, data := range tests {
		encoded := Base85Encode(data)
		decoded, err := Base85Decode(encoded)
		if err != nil {
			t.Errorf("Base85Decode(%q) error: %v", encoded, err)
			continue
		}
		if !bytes.Equal(data, decoded) {
			t.Errorf("Base85 roundtrip: got %v, want %v", decoded, data)
		}
	}
}

func TestConvert(t *testing.T) {
	// hex -> base64
	result, err := Convert("68656c6c6f", EncodingHex, EncodingBase64)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if result != "aGVsbG8=" {
		t.Errorf("Convert hex->base64: got %q, want %q", result, "aGVsbG8=")
	}

	// base64 -> hex
	result, err = Convert("aGVsbG8=", EncodingBase64, EncodingHex)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if result != "68656c6c6f" {
		t.Errorf("Convert base64->hex: got %q, want %q", result, "68656c6c6f")
	}
}

func TestDetect(t *testing.T) {
	// base64
	candidates := Detect("aGVsbG8=")
	found := false
	for _, c := range candidates {
		if c == EncodingBase64 {
			found = true
		}
	}
	if !found {
		t.Errorf("Detect should find base64 for 'aGVsbG8=': got %v", candidates)
	}

	// hex
	candidates = Detect("68656c6c6f")
	found = false
	for _, c := range candidates {
		if c == EncodingHex {
			found = true
		}
	}
	if !found {
		t.Errorf("Detect should find hex for '68656c6c6f': got %v", candidates)
	}

	// base58
	candidates = Detect("2NEpo7TZRRrLZSi2U")
	found = false
	for _, c := range candidates {
		if c == EncodingBase58 {
			found = true
		}
	}
	if !found {
		t.Errorf("Detect should find base58 for '2NEpo7TZRRrLZSi2U': got %v", candidates)
	}
}

func TestUnsupportedEncoding(t *testing.T) {
	_, err := Encode([]byte("test"), Encoding("invalid"))
	if err != ErrUnsupportedEncoding {
		t.Errorf("expected ErrUnsupportedEncoding, got %v", err)
	}

	_, err = Decode("test", Encoding("invalid"))
	if err != ErrUnsupportedEncoding {
		t.Errorf("expected ErrUnsupportedEncoding, got %v", err)
	}
}

func TestAllEncodings(t *testing.T) {
	encodings := AllEncodings()
	if len(encodings) != 8 {
		t.Errorf("expected 8 encodings, got %d", len(encodings))
	}
}
