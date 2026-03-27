package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func AESEncrypt(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	iv := make([]byte, aes.BlockSize)
	for i := range iv {
		iv[i] = '0'
	}
	cbc := cipher.NewCBCEncrypter(block, iv)
	padded := pkcs7Padding([]byte(plaintext), aes.BlockSize)
	encrypted := make([]byte, len(padded))
	cbc.CryptBlocks(encrypted, padded)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func AESDecrypt(key []byte, ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	iv := make([]byte, aes.BlockSize)
	for i := range iv {
		iv[i] = '0'
	}
	cbc := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	cbc.CryptBlocks(decrypted, data)
	unpadded := pkcs7Unpadding(decrypted)
	return string(unpadded), nil
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

func pkcs7Unpadding(data []byte) []byte {
	length := len(data)
	padding := int(data[length-1])
	return data[:length-padding]
}
