package secmsg

import (
	"errors"
	"time"

	"github.com/fernet/fernet-go"
)

const secretKeyLength = 32

var (
	ErrSecretKeyLength    = errors.New("secret key too short")
	ErrVerificationFailed = errors.New("message verification failed")
)

type MessageService struct {
	encKey  *fernet.Key
	decKeys []*fernet.Key
}

// NewMessageService sets up a message service with the given key.
func NewMessageService(secret []byte) (*MessageService, error) {
	// accept a byte slice so we can change the crypto more easily later on
	if len(secret) < secretKeyLength {
		return nil, ErrSecretKeyLength
	}

	var key fernet.Key
	copy(key[:], secret[:32])

	return &MessageService{
		encKey:  &key,
		decKeys: []*fernet.Key{&key},
	}, nil
}

// Encode encodes a message, returning an error on failure.
func (svc *MessageService) Encode(msg []byte) ([]byte, error) {
	return fernet.EncryptAndSign(msg, svc.encKey)
}

// MustEncode encodes a message, calling panic on failure.
func (svc *MessageService) MustEncode(msg []byte) []byte {
	token, err := svc.Encode(msg)
	if err != nil {
		panic(err)
	}
	return token
}

// Verify verifies and decodes a token, returning an error on failure.
func (svc *MessageService) Decode(token []byte, maxage time.Duration) ([]byte, error) {
	msg := fernet.VerifyAndDecrypt(token, maxage, svc.decKeys)
	if msg == nil {
		return nil, ErrVerificationFailed
	}
	return msg, nil
}

// Decode verifies and decodes a token, returning an empty slice on failure.
func (svc *MessageService) MustDecode(token []byte, maxage time.Duration) []byte {
	return fernet.VerifyAndDecrypt(token, maxage, svc.decKeys)
}
