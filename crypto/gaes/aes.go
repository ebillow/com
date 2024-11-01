package gaes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

func NewEncrypter(key []byte, iv []byte) (cipher.BlockMode, error) {
	block, err := aes.NewCipher(key)
	// 判断是否创建成功
	if err != nil {
		panic(err)
	}

	// 创建一个密码分组为链接模式的, 底层使用DES加密的BlockMode接口
	return cipher.NewCBCEncrypter(block, iv), nil
}

func NewDecrypter(key []byte, iv []byte) (cipher.BlockMode, error) {
	// 创建并返回一个使用DES算法的cipher.Block接口
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	// 创建一个密码分组为链接模式的, 底层使用DES解密的BlockMode接口
	return cipher.NewCBCDecrypter(block, iv), nil
}

// 加密函数
func EnCrypt(src []byte, blockMode cipher.BlockMode) []byte {
	// 明文组数据填充
	paddingText := PKCS7Padding(src, blockMode.BlockSize())
	// 加密
	//dst := make([]byte, len(paddingText))
	blockMode.CryptBlocks(paddingText, paddingText)
	return paddingText
}

func DeCrypt(src []byte, blockMode cipher.BlockMode) []byte {
	// 解密
	//dst := make([]byte, len(src))
	blockMode.CryptBlocks(src, src)
	// 分组移除
	return PKCS7UnPadding(src)
}

// PKCS7Padding 填充
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(plaintext []byte) []byte {
	if len(plaintext) == 0 {
		return nil
	}
	length := len(plaintext)
	unpadding := int(plaintext[length-1])
	return plaintext[:length-unpadding]
}
