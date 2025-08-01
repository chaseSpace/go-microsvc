package ucrypto

import (
	"microsvc/util/urand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCryptoAes(t *testing.T) {
	// invalid keys
	for _, k := range []string{"", "1", "15_chars_key_xx", "17_chars_key_xxxx"} {
		assert.Panics(t, func() {
			NewCryptoAes(k)
		}, ErrInvalidAesKey.Error())
	}

	// right keys
	for _, k := range []string{"16_chars_key_xxx", "24_chars_key_xxxxxxxxxxx", "32_chars_key_xxxxxxxxxxxxxxxxxxx"} {
		assert.NotPanics(t, func() {
			NewCryptoAes(k)
		})
	}
}

func TestCryptoAes_Encrypt(t *testing.T) {
	for _, key := range []string{"16_chars_key_xxx", "24_chars_key_xxxxxxxxxxx", "32_chars_key_xxxxxxxxxxxxxxxxxxx"} {

		for i := 0; i < 100; i++ {
			c := NewCryptoAes(key)

			src := []byte(urand.Strings(i))
			println(string(src))
			b, err := c.Encrypt(src)
			assert.Nil(t, err)
			assert.NotNil(t, b)

			src2, err := c.Decrypt(b)
			assert.Nil(t, err)
			assert.Equal(t, src, src2)
		}
	}
}
