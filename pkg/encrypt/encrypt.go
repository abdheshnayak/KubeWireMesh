package encrypt

import (
	"fmt"
	"math/rand"

	"golang.org/x/crypto/nacl/box"
)

func Encrypt(data []byte, publicKey, privateKey [32]byte) ([]byte, error) {

	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}

	encrypted := box.Seal(nonce[:], data, &nonce, &publicKey, &privateKey)

	return encrypted, nil
}

func Decrypt(encrypted []byte, publicKey, privateKey [32]byte) ([]byte, error) {

	var nonce [24]byte
	copy(nonce[:], encrypted[:24])

	decrypted, ok := box.Open(nil, encrypted[24:], &nonce, &publicKey, &privateKey)
	if !ok {
		return nil, fmt.Errorf("failed to decrypt")
	}

	return decrypted, nil
}
