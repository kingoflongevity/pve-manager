package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret JWT 签名密钥，生产环境应通过环境变量配置
var jwtSecret = []byte("pve-manager-jwt-secret-change-in-production")

// aesKey AES-256 加密密钥，用于加密存储的 PVE 密码
var aesKey = []byte("pve-mgr-aes-key-32-bytes-long!!!")

// Claims JWT token 负载
type Claims struct {
	Username string `json:"username"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Realm    string `json:"realm"`
	Creds    string `json:"creds"` // AES 加密后的 PVE 密码
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT token
// 包含用户信息和加密的 PVE 密码，有效期 7 天
func GenerateJWT(username, host string, port int, realm, creds string) (string, int64, error) {
	expiresIn := int64(7 * 24 * 3600) // 7天
	now := time.Now()
	claims := Claims{
		Username: username,
		Host:     host,
		Port:     port,
		Realm:    realm,
		Creds:    creds,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiresIn) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "pve-manager",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", 0, err
	}
	return signedToken, expiresIn, nil
}

// ParseJWT 解析并验证 JWT token
func ParseJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// EncryptPassword 使用 AES-256-GCM 加密密码
func EncryptPassword(password string) (string, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword 使用 AES-256-GCM 解密密码
func DecryptPassword(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文过短")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
