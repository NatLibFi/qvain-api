// Package randomkey generates cryptographically secure keys of lengths 16, 32 or 64.
package randomkey

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

type Key []byte

func (k Key) Bytes() []byte {
	return k
}

func (k Key) Hex() string {
	return hex.EncodeToString(k)
}

func FromHex(s string) (*Key, error) {
	key, err := hex.DecodeString(s)
	return (*Key)(&key), err
}

func (k Key) Base64() string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(k)
}

func FromBase64(s string) (*Key, error) {
	key, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(s)
	return (*Key)(&key), err
}

func (k Key) Len() int {
	return len(k)
}

func (k Key) String() string {
	return hex.EncodeToString(k)
}

func (k Key) GoString() string {
	//return fmt.Sprintf("Key(0x%s)", k.Hex())
	//return fmt.Sprintf("Key(0x %x)", k)
	return "Key(" + k.Hex() + ")"
}

func genRandomBytes(n int) (*Key, error) {
	var key Key = make([]byte, n)

	_, err := rand.Read(key)
	if err != nil {
		return &Key{}, fmt.Errorf("error generating random key: %s", err)
	}
	return &key, nil
}

func Random8() (*Key, error) {
	return genRandomBytes(8)
}

func Random16() (*Key, error) {
	return genRandomBytes(16)
}

func Random32() (*Key, error) {
	return genRandomBytes(32)
}

func Random64() (*Key, error) {
	return genRandomBytes(64)
}
