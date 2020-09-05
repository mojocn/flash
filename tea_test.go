package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/tea"
	"golang.org/x/crypto/xtea"
	"io"
	"testing"
)

func TestTeaDemo(t *testing.T) {

	key := []byte("mojotv.cn.=.good") //长度必须为16byte
	c, err := tea.NewCipherWithRounds(key, 8)
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte("mojotvcn") //长度必须为8byte
	dst := make([]byte, 8)    //长度必须为8byte
	c.Encrypt(dst, raw)
	raw2 := make([]byte, 8) //长度必须为8byte
	c.Decrypt(raw2, dst[:])

	if !bytes.Equal(raw, raw2) {
		t.Error("失败")
	}
	t.Log("验证成功")
}
func TestXteaDemo(t *testing.T) {

	key := []byte("mojotv.cn.=.good") //长度必须为16byte
	c, err := xtea.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte("mojotvcn") //长度必须为8byte
	dst := make([]byte, 8)    //长度必须为8byte
	c.Encrypt(dst, raw)
	raw2 := make([]byte, 8) //长度必须为8byte
	c.Decrypt(raw2, dst[:])

	if !bytes.Equal(raw, raw2) {
		t.Error("失败")
	}
	t.Log("xtea验证成功")
}

//使用PKCS7进行填充
func pKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//使用PKCS7进行填充 复原
func pKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//XteaCbcEncrypt key 长度必须为16 byte
func XteaCbcEncrypt(rawData, key []byte) ([]byte, error) {
	block, err := xtea.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("key 只能为16bytes, %v", err)
	}

	//填充原文
	blockSize := block.BlockSize()
	rawData = pKCS7Padding(rawData, blockSize)
	//初始向量IV必须是唯一，但不需要保密
	cipherText := make([]byte, blockSize+len(rawData))
	//block大小 16
	iv := cipherText[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	//block大小和初始向量大小一定要一致
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[blockSize:], rawData)

	return cipherText, nil
}

//XteaCbcDecrypt key 长度必须为16 byte
func XteaCbcDecrypt(encryptData, key []byte) ([]byte, error) {
	block, err := xtea.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("key 只能为16bytes, %v", err)
	}

	blockSize := block.BlockSize()

	if len(encryptData) < blockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := encryptData[:blockSize]
	encryptData = encryptData[blockSize:]

	// CBC mode always works in whole blocks.
	if len(encryptData)%blockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(encryptData, encryptData)
	//解填充
	encryptData = pKCS7UnPadding(encryptData)
	return encryptData, nil
}

func XteaB64urlEncrypt(rawData, key []byte) (string, error) {
	data, err := XteaCbcEncrypt(rawData, key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func XteaB64urlDecrypt(rawData string, key []byte) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(rawData)
	if err != nil {
		return "", err
	}
	dnData, err := XteaCbcDecrypt(data, key)
	if err != nil {
		return "", err
	}
	return string(dnData), nil
}

func TestXteaCbcB64url(t *testing.T) {
	key := []byte("mojotv.cn.=.good") //长度必须为16byte
	raw := "mojotv.cn and golang are great friends"
	ciper, err := XteaB64urlEncrypt([]byte(raw), key)
	if err != nil {
		t.Error("xtea cbc base64 url 加密失败", err)
		return
	}
	decrypt, err := XteaB64urlDecrypt(ciper, key)
	if err != nil {
		t.Error("xtea cbc base64 url 解密失败", err)
		return
	}
	if decrypt != raw {
		t.Error("解密结果不正确")
	}
}
