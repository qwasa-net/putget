package putget

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

func hashKey(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

func encrypt(content []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher(hashKey(key))
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(content))
	stream.XORKeyStream(ciphertext, content)

	return append(iv, ciphertext...), nil
}

func decrypt(ciphertext []byte, key string) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	data := ciphertext[aes.BlockSize:]

	block, err := aes.NewCipher(hashKey(key))
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(data))
	stream.XORKeyStream(plaintext, data)

	return plaintext, nil
}
