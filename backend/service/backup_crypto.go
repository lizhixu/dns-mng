package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"crypto/sha256"
)

const (
	// pbkdf2Iterations 控制密钥派生的迭代次数，200k 在安全性与速度之间取得平衡。
	pbkdf2Iterations = 200_000
	keyLen           = 32 // AES-256
	saltLen          = 32
	nonceLen         = 12 // AES-GCM 标准 nonce 长度
)

// encryptedEnvelope 是加密后的备份文件结构。
type encryptedEnvelope struct {
	Version   int              `json:"version"`
	Encrypted bool             `json:"encrypted"`
	KDF       kdfParams        `json:"kdf"`
	Nonce     string           `json:"nonce"`
	Ciphertext string          `json:"ciphertext"`
}

type kdfParams struct {
	Salt string `json:"salt"`
	Iter int    `json:"iter"`
}

// deriveKey 从密码和 salt 派生 AES-256 密钥。
func deriveKey(password string, salt []byte) ([]byte, error) {
	return pbkdf2.Key(sha256.New, password, salt, pbkdf2Iterations, keyLen)
}

// EncryptBackup 将明文 JSON 使用 AES-256-GCM 加密，返回完整的加密 JSON 文件字节。
func EncryptBackup(plainJSON []byte, password string) ([]byte, error) {
	if password == "" {
		return plainJSON, nil
	}

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	key, err := deriveKey(password, salt)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plainJSON, nil)

	envelope := encryptedEnvelope{
		Version:    1,
		Encrypted:  true,
		KDF:        kdfParams{Salt: hex.EncodeToString(salt), Iter: pbkdf2Iterations},
		Nonce:      hex.EncodeToString(nonce),
		Ciphertext: hex.EncodeToString(ciphertext),
	}

	return json.MarshalIndent(envelope, "", "  ")
}

// DecryptBackup 解密加密的备份文件，返回原始明文 JSON 字节。
func DecryptBackup(fileBytes []byte, password string) ([]byte, error) {
	var envelope encryptedEnvelope
	if err := json.Unmarshal(fileBytes, &envelope); err != nil {
		return nil, fmt.Errorf("parse backup file: %w", err)
	}

	if !envelope.Encrypted {
		// 未加密文件，直接返回
		return fileBytes, nil
	}

	if password == "" {
		return nil, fmt.Errorf("此备份文件已加密，请输入加密密码")
	}

		salt, err := hex.DecodeString(envelope.KDF.Salt)
		if err != nil {
			return nil, fmt.Errorf("decode salt: %w", err)
		}

		key, err := deriveKey(password, salt)
		if err != nil {
			return nil, fmt.Errorf("derive key: %w", err)
		}

		block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce, err := hex.DecodeString(envelope.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}

	ciphertext, err := hex.DecodeString(envelope.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("解密失败，密码可能不正确")
	}

	return plain, nil
}
