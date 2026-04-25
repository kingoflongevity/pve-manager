package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// AESCipher AES-256-GCM 加密器
// 使用 GCM 模式提供认证加密，确保数据完整性和机密性
type AESCipher struct {
	gcm cipher.AEAD
}

// NewAESCipher 创建 AES-256-GCM 加密器
// key 必须是 32 字节的密钥，否则返回错误
// GCM 模式不需要 IV 管理，每次加密自动生成随机 nonce
func NewAESCipher(key []byte) (*AESCipher, error) {
	if len(key) != 32 {
		return nil, errors.New("AES-256 密钥长度必须为 32 字节")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AESCipher{gcm: gcm}, nil
}

// Encrypt 加密明文数据
// 返回 base64 编码的密文（包含 nonce）
// 每次加密都会生成随机 nonce 确保安全性
func (a *AESCipher) Encrypt(plaintext []byte) (string, error) {
	// 生成随机 nonce（GCM 默认 12 字节）
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密并附加认证标签
	ciphertext := a.gcm.Seal(nonce, nonce, plaintext, nil)

	// 返回 base64 编码结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密密文数据
// 输入为 base64 编码的密文（包含 nonce）
// 自动提取 nonce 并验证数据完整性
func (a *AESCipher) Decrypt(encodedText string) ([]byte, error) {
	// 解码 base64
	ciphertext, err := base64.StdEncoding.DecodeString(encodedText)
	if err != nil {
		return nil, err
	}

	nonceSize := a.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("密文长度不足")
	}

	// 提取 nonce 和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密并验证
	plaintext, err := a.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptString 加密字符串并返回 base64 编码
// 便捷方法，直接处理字符串类型
func (a *AESCipher) EncryptString(plaintext string) (string, error) {
	return a.Encrypt([]byte(plaintext))
}

// DecryptString 解密并返回字符串
// 便捷方法，直接返回字符串类型
func (a *AESCipher) DecryptString(encodedText string) (string, error) {
	plaintext, err := a.Decrypt(encodedText)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// GenerateAESKey 生成随机 AES-256 密钥
// 返回 32 字节的随机密钥，可用于初始化加密器
func GenerateAESKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}
