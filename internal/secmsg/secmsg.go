package secmsg

import (
	"errors"
	"time"

	"github.com/fernet/fernet-go"
)

const secretKeyLength = 32

var (
	ErrSecretKeyLength    = errors.New("secret key too short")
	ErrVerificationFailed = errors.New("verification failed")
)

type TokenService struct {
	encKey  *fernet.Key
	decKeys []*fernet.Key
}

// NewTokenService sets up a token service with the given key.
func NewTokenService(secret []byte) (*TokenService, error) {
	// accept a byte slice so we can change the crypto more easily later on
	if len(secret) < secretKeyLength {
		return nil, ErrSecretKeyLength
	}

	var key fernet.Key
	copy(key[:], secret[:32])

	return &TokenService{
		encKey:  &key,
		decKeys: []*fernet.Key{&key},
	}, nil
}

// Encode encodes a message, returning an error on failure.
func (svc *TokenService) Encode(msg []byte) ([]byte, error) {
	return fernet.EncryptAndSign(msg, svc.encKey)
}

// MustEncode encodes a message, calling panic on failure.
func (svc *TokenService) MustEncode(msg []byte) []byte {
	token, err := svc.Encode(msg)
	if err != nil {
		panic(err)
	}
	return token
}

// Verify verifies and decodes a token, returning an error on failure.
func (svc *TokenService) Decode(token []byte, maxage time.Duration) ([]byte, error) {
	msg := fernet.VerifyAndDecrypt(token, maxage, svc.decKeys)
	if msg == nil {
		return nil, ErrVerificationFailed
	}
	return msg, nil
}

// Decode verifies and decodes a token, returning an empty slice on failure.
func (svc *TokenService) MustDecode(token []byte, maxage time.Duration) []byte {
	return fernet.VerifyAndDecrypt(token, maxage, svc.decKeys)
}
