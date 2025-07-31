package ucrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

type CryptoAes struct {
	key []byte
}

var (
	ErrInvalidAesKey = errors.New("密钥长度必须为16、24或32字节")
)

// NewCryptoAes 创建一个 CryptoAes 实例，并设置密钥
func NewCryptoAes(key string) *CryptoAes {
	keyBytes := []byte(key)
	if len(keyBytes) != 16 && len(keyBytes) != 24 && len(keyBytes) != 32 {
		panic(ErrInvalidAesKey)
	}

	return &CryptoAes{key: keyBytes}
}

// PKCS7Padding Padding for plaintext to be a multiple of the block size
func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7UnPadding Unpadding after decryption
func PKCS7UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

func (c *CryptoAes) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	// Pad the plaintext to be a multiple of the block size
	plaintext = PKCS7Padding(plaintext, aes.BlockSize)

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 采用CBC模式
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *CryptoAes) Decrypt(ciphertextBase64 string) ([]byte, error) {
	// Decode the base64-encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext length is not a multiple of the block size")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Remove padding
	plaintext := PKCS7UnPadding(ciphertext)
	return plaintext, nil
}

func Sha1(input string, salt ...string) (string, error) {
	saltedInput := strings.Join(append(salt, input), "")
	h := sha1.New()
	h.Write([]byte(saltedInput))
	hash := h.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}

func MustSha1(input string, salt ...string) string {
	hash, _ := Sha1(input, salt...)
	return hash
}

func Md5(input string) string {
	return fmt.Sprintf("%x", md5.New().Sum([]byte(input)))
}
