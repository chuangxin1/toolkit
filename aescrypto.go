package toolkit

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

var (
	key string
)

// AesCrypto define
type AesCrypto struct {
	Key []byte
}

// SetAesCryptoKey set key,
// key长度：16, 24, 32 bytes 对应 AES-128, AES-192, AES-256
func SetAesCryptoKey(password string) {
	key = password
}

// GetAesCryptoKey get current key
func GetAesCryptoKey() string {
	return key
}

// NewAesCrypto new AesCrypto
func NewAesCrypto() *AesCrypto {
	fmt.Println(key)
	return &AesCrypto{[]byte(key)}
}

// SetKey set key
func (a *AesCrypto) SetKey(key string) {
	a.Key = []byte(key)
}

// Encrypt encrypt data
func (a *AesCrypto) Encrypt(origData []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.Key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, a.Key[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// Decrypt decrypt data
func (a *AesCrypto) Decrypt(crypted []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.Key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, a.Key[:blockSize])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(data []byte) []byte {
	length := len(data)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}
