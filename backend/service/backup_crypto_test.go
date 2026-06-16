package service

import (
	"encoding/json"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	original := map[string]interface{}{
		"version":     1,
		"exported_at": "2026-06-17T10:00:00Z",
		"encrypted":   false,
		"data": map[string]interface{}{
			"accounts": []map[string]string{
				{"name": "My CF", "provider_type": "cloudflare", "api_key": "secret123"},
			},
		},
	}

	plainJSON, _ := json.MarshalIndent(original, "", "  ")

	// --- 无密码（不加密） ---
	encrypted, err := EncryptBackup(plainJSON, "")
	if err != nil {
		t.Fatalf("EncryptBackup (no password): %v", err)
	}
	decrypted, err := DecryptBackup(encrypted, "")
	if err != nil {
		t.Fatalf("DecryptBackup (no password): %v", err)
	}
	if string(decrypted) != string(plainJSON) {
		t.Error("no-password roundtrip mismatch")
	}

	// --- 有密码（加密） ---
	password := "test-password-123"
	encrypted, err = EncryptBackup(plainJSON, password)
	if err != nil {
		t.Fatalf("EncryptBackup: %v", err)
	}

	// 验证结果是 JSON 且 encrypted=true
	var envelope map[string]interface{}
	if err := json.Unmarshal(encrypted, &envelope); err != nil {
		t.Fatalf("encrypted output not JSON: %v", err)
	}
	if envelope["encrypted"] != true {
		t.Error("expected encrypted=true")
	}

	// 正确密码解密
	decrypted, err = DecryptBackup(encrypted, password)
	if err != nil {
		t.Fatalf("DecryptBackup (correct password): %v", err)
	}
	if string(decrypted) != string(plainJSON) {
		t.Error("encrypted roundtrip mismatch")
	}

	// 错误密码应失败
	_, err = DecryptBackup(encrypted, "wrong-password")
	if err == nil {
		t.Error("expected error for wrong password")
	}

	// 加密文件不传密码应失败
	_, err = DecryptBackup(encrypted, "")
	if err == nil {
		t.Error("expected error when password empty for encrypted file")
	}
}
