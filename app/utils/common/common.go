package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"
)

func IsValidEmail(email string) bool {
	// 定义邮箱地址的正则表达式
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// 编译正则表达式
	regexpPattern, err := regexp.Compile(emailRegex)
	if err != nil {
		panic(err)
	}

	// 使用正则表达式匹配邮箱地址
	return regexpPattern.MatchString(email)
}

func init() {
	assertAvailablePRNG()
}

func assertAvailablePRNG() {
	// Assert that a cryptographically secure PRNG is available.
	// Panic otherwise.
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomStringURLSafe(n int) (string, error) {
	neededBytes := (n*6 + 7) / 8 // 使用 6/8 而不是 3/4，以避免填充字符
	b, err := GenerateRandomBytes(neededBytes)
	if err != nil {
		return "", err
	}

	result := base64.RawURLEncoding.EncodeToString(b)

	// 截取结果字符串的前 n 个字符，确保最终长度为 n
	if len(result) > n {
		result = result[:n]
	}

	return result, nil
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func ReplaceEndString(input, pattern, repl string) (string, error) {
	regexPattern := pattern + "$" // 在正则表达式末尾添加 $，表示匹配末尾
	r, err := regexp.Compile(regexPattern)
	if err != nil {
		return "", err
	}

	result := r.ReplaceAllString(input, repl)
	return result, nil
}

// 判断是否包含数字
func ContainsInt(target int64, slice []int64) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

// 判断是否包含字符串
func ContainsString(target string, slice []string) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

// 将结构序列化字符串
func SeriToString(data interface{}) string {
	// 使用 json.Marshal 将数据序列化为 JSON 字符串
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	// 将字节切片转换为字符串
	return string(jsonBytes)
}

func SeriToJsonRawMessage(data interface{}) *json.RawMessage {
	tmp := SeriToString(data)
	result := json.RawMessage([]byte(tmp))

	return &result
}

// generateHMAC 生成 HMAC SHA256 签名
func GenerateHMAC(key, message string) string {
	// 将密钥转换为字节数组
	keyBytes := []byte(key)

	// 创建一个新的 HMAC 实例，使用 SHA256 散列算法
	h := hmac.New(sha256.New, keyBytes)

	// 写入要签名的消息
	h.Write([]byte(message))

	// 计算签名
	signature := h.Sum(nil)

	// 使用 Base64 编码签名结果
	return base64.StdEncoding.EncodeToString(signature)
}

func AbsInt(num int64) int64 {
	if num < 0 {
		return -num
	}
	return num
}

func IntJoin(nums []int64, sep string) string {
	var builder strings.Builder

	for i, num := range nums {
		if i > 0 {
			builder.WriteString(sep)
		}
		builder.WriteString(fmt.Sprintf("%d", num))
	}

	return builder.String()
}

var reStripHTML = regexp.MustCompile(`<[^>]*>|[\s\r\n]+`)

// 使用正则表达式匹配并替换 HTML 标签、空格和换行符为空字符串
func StripHTML(input string) string {
	return reStripHTML.ReplaceAllString(input, "")
}

func GetSummary(text string, maxLength int) string {
	str := StripHTML(text)
	if len(str) <= maxLength {
		return str
	}
	return str[:maxLength] + "..."
}

// AES加密
func EncryptAES(dataStr string, keyStr string) (string, error) {

	data := []byte(dataStr)
	key := []byte(keyStr)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cipher.NewCFBEncrypter(block, iv).XORKeyStream(ciphertext[aes.BlockSize:], data)

	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return encoded, nil
}

// AES解密
func DecryptAES(ciphertextStr, keyStr string) (string, error) {

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextStr)
	if err != nil {
		return "", fmt.Errorf("解密失败！%s,%v", ciphertextStr, err)
	}

	key := []byte(keyStr)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext is too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	cipher.NewCFBDecrypter(block, iv).XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
