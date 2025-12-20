package common

import (
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestGenerateHMAC(t *testing.T) {
	// 你的密钥
	secretKey := "" // variable.ConfigYml.GetString("EZBase.SecretKey")

	// 获取当前毫秒级时间戳
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	timestampStr := fmt.Sprintf("%d", timestamp)

	// 生成 HMAC SHA256 签名
	signature := GenerateHMAC(secretKey, timestampStr)
	fmt.Println("signature:", signature, "timestampStr:", timestampStr)

	// 创建一个 decimal 对象表示金额
	amount := decimal.NewFromFloat(1234)

	// 使用 decimal 对象的 StringFixed 方法格式化金额并保留两位小数
	formattedAmount := amount.StringFixed(2)

	// 打印结果
	fmt.Println("Formatted Amount:", formattedAmount)
}

func TestAES(t *testing.T) {
	data := "5345345"
	key := "8da3e5c3cba5e80123ce260b2c06a18f"
	// 加密
	encryptedData, err := EncryptAES(data, key)
	if err != nil {
		t.Fatalf("Encryption error: %v", err)
		return
	}

	fmt.Printf("Encrypted: %x\n", encryptedData)

	// 解密
	decryptedData, err := DecryptAES(encryptedData, key)
	if err != nil {
		t.Fatalf("Encryption error: %v", err)
		return
	}

	fmt.Println("Decrypted:", string(decryptedData))
}
