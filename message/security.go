package message

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/sha3"
	"io"
	"strconv"
	// "reflect"
)

var orgKey = []byte("yao32bytes nameduo zenmexie aaaa") //32bytes
var cfgKey = "hello key"

func Sign(data, key []byte) ([]byte, error) {
	hash := sha3.Sum256(data)
	return AesEn(hash[:], key)
}
func Verify(recv, data, key []byte) bool {
	recvHash, err := AesDe(recv, key)
	if err != nil {
		return false
	}
	hash := sha3.Sum256(data)
	return bytes.Equal(recvHash, hash[:])
	// return reflect.DeepEqual(recvHash, hash[:])
}

func NewKey1(n uint64) []byte {
	newKey := make([]byte, 32)
	binary.LittleEndian.PutUint64(newKey[:8], n^binary.LittleEndian.Uint64(orgKey[:8]))
	binary.LittleEndian.PutUint64(newKey[8:16], n^binary.LittleEndian.Uint64(orgKey[8:16]))
	binary.LittleEndian.PutUint64(newKey[16:24], n^binary.LittleEndian.Uint64(orgKey[16:24]))
	binary.LittleEndian.PutUint64(newKey[24:], n^binary.LittleEndian.Uint64(orgKey[24:]))
	return newKey
}
func NewKey(n uint64) []byte {
	newKey := sha3.Sum256([]byte(cfgKey + strconv.FormatUint(n, 10)))
	return newKey[:]
}

func AesEn(data, key []byte) ([]byte, error) {
	plaintext := padding(data, aes.BlockSize)
	block, err := aes.NewCipher(key)
	if err != nil {
		// flog.LogFile.Fatal(err)
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}
func AesDe(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		// flog.LogFile.Panic(err)
		return nil, err
	}

	if len(data) < aes.BlockSize {
		return nil, errors.New("ciphertext to short")
	}
	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	// var plaintext []byte
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	return unpadding(plaintext), nil
}

//PKCS7
func padding(src []byte, blockSize int) []byte {
	pad := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(pad)}, pad)
	return append(src, padtext...)
}
func unpadding(src []byte) []byte {
	length := len(src)
	delLen := int(src[length-1])
	return src[:(length - delLen)]
}
