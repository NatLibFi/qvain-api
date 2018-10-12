package secmsg

import (
	"encoding/hex"
	"testing"
	"time"
)

const encodedSecretKey = "b290a87f12ee82966a868233784bf56704a74deaa95e381b07220807945e31d1"

var secretKey = func() []byte {
	key, err := hex.DecodeString(encodedSecretKey)
	if err != nil {
		panic(err)
	}
	return key
}()

var tests = []struct {
	name      string
	cleartext []byte
}{
	{
		name:      "empty string",
		cleartext: []byte(""),
	},
	{
		name:      "hello world",
		cleartext: []byte("Hello, World!"),
	},
	{
		name:      "binary",
		cleartext: []byte("\xFFinvalid unicode\xFF\xFE\nmäh€\x01\x02\x03!"),
	},
	{
		// length: 106B
		name:      "json",
		cleartext: []byte(`{"id":666,"name":"Jack Fullstack","nested":{"much structure":"such deep","so data":["very","impressive"]}}`),
	},
	{
		// length: 85B (same value as used for json test)
		name:      "msgpack",
		cleartext: []byte{0x83, 0xa2, 0x69, 0x64, 0xd1, 0x2, 0x9a, 0xa4, 0x6e, 0x61, 0x6d, 0x65, 0xae, 0x4a, 0x61, 0x63, 0x6b, 0x20, 0x46, 0x75, 0x6c, 0x6c, 0x73, 0x74, 0x61, 0x63, 0x6b, 0xa6, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x82, 0xae, 0x6d, 0x75, 0x63, 0x68, 0x20, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x75, 0x72, 0x65, 0xa9, 0x73, 0x75, 0x63, 0x68, 0x20, 0x64, 0x65, 0x65, 0x70, 0xa7, 0x73, 0x6f, 0x20, 0x64, 0x61, 0x74, 0x61, 0x92, 0xa4, 0x76, 0x65, 0x72, 0x79, 0xaa, 0x69, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x76, 0x65},
	},
}

func TestEncodeDecode(t *testing.T) {
	svc, err := NewMessageService(secretKey)
	if err != nil {
		t.Fatal(err)
	}

	maxage := 300 * time.Second

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token, err := svc.Encode(test.cleartext)
			if err != nil {
				t.Error("encoding error:", err)
			}
			cleartext := svc.MustDecode(token, maxage)
			if string(cleartext) != string(test.cleartext) {
				t.Errorf("encoding error: expected %q, got %q", test.cleartext, cleartext)
			}
		})
	}
}

func TestExpiredMessage(t *testing.T) {
	svc, err := NewMessageService(secretKey)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("hello")
	maxage := 500 * time.Millisecond

	token := svc.MustEncode(msg)
	time.Sleep(1 * time.Second)
	cleartext, err := svc.Decode(token, maxage)
	if err != ErrVerificationFailed {
		t.Errorf("expected ErrVerificationFailed, got: %v", err)
	}
	if string(cleartext) != "" {
		t.Errorf("expected cleartext not to leak, but got: %q", cleartext)
	}
}

func TestNeverExpiringMessage(t *testing.T) {
	svc, err := NewMessageService(secretKey)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("hello")
	maxage := -1 * time.Second

	token := svc.MustEncode(msg)
	time.Sleep(1 * time.Second)
	cleartext, err := svc.Decode(token, maxage)
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if string(cleartext) != "hello" {
		t.Errorf("encoding error: expected %q, got %q", "hello", string(cleartext))
	}
}

// nil message returns a "non untyped nil" empty slice after decoding
func TestNilClearText(t *testing.T) {
	svc, err := NewMessageService(secretKey)
	if err != nil {
		t.Fatal(err)
	}

	var msg []byte
	maxage := 5 * time.Second

	token := svc.MustEncode(msg)
	cleartext, err := svc.Decode(token, maxage)
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if string(cleartext) != "" {
		t.Errorf("decoding error: expected %q, got %q", "", string(cleartext))
	}
	if cleartext == nil {
		t.Errorf("decoding error: expected %#v, got %#v", []byte{}, cleartext)
	}
	if len(cleartext) != 0 {
		t.Errorf("decoding error: expected length %d, got %d", 0, len(cleartext))
	}
}

func TestTamperedMessage(t *testing.T) {
	svc, err := NewMessageService(secretKey)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("hello")
	maxage := 300 * time.Second

	token := svc.MustEncode(msg)
	cleartext, err := svc.Decode(token[:len(token)-1], maxage)
	if err != ErrVerificationFailed {
		t.Errorf("expected ErrVerificationFailed, got: %v", err)
	}
	if string(cleartext) != "" {
		t.Errorf("expected cleartext not to leak, but got: %q", cleartext)
	}
}

func TestTooShortKey(t *testing.T) {
	_, err := NewMessageService(secretKey[0:31])
	if err != ErrSecretKeyLength {
		t.Error(err)
	}
}

func BenchmarkEncode(b *testing.B) {
	svc, err := NewMessageService(secretKey)
	if err != nil {
		b.Fatal(err)
	}

	for _, test := range tests {
		clearBytes := []byte(test.cleartext)
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = svc.MustEncode(clearBytes)
			}
		})
	}
}

func BenchmarkDecode(b *testing.B) {
	svc, err := NewMessageService(secretKey)
	if err != nil {
		b.Fatal(err)
	}

	for _, test := range tests {
		token := svc.MustEncode([]byte(test.cleartext))
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = svc.MustDecode(token, -1)
			}
		})
	}
}
