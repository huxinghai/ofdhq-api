package douyin

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"ofdhq-api/app/global/variable"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 抖音本地生活 SPI 回调签名校验，规则见：
// https://partner.open-douyin.com/docs/resource/zh-CN/local-life/develop/preparation/spi-signature-rules
//
// 待签名明文 str1 的拼装规则：
//  1. str1 = client_secret + "&"
//  2. 取 URL 上除 sign（大小写不敏感）以外的全部 query 参数，按 key 字典升序排列；
//     同一个 key 有多个 value 时 value 也升序排列，每 组追加 "&key=value"
//  3. body 不为空时，末尾追加 "&http_body=" + 原始 body 字符串
//
// 新版签名：header x-life-sign = SHA-256(str1) 的小写 hex
// 旧版签名：URL 参数 sign    = MD5(str1)     的小写 hex
//
// 注意：签名针对的是抖音发来的原始 body 字节，必须先读原始 body 再验签，
// 严禁先 json.Unmarshal 再重新序列化（字段顺序/转义变化会导致验签失败）。

// BuildSignString 按上述规则拼接待签名明文 str1
func BuildSignString(clientSecret string, query url.Values, body []byte) string {
	var sb strings.Builder
	sb.WriteString(clientSecret)
	sb.WriteString("&")

	keys := make([]string, 0, len(query))
	for k := range query {
		if strings.EqualFold(k, "sign") {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		values := append([]string(nil), query[k]...)
		sort.Strings(values)
		for _, v := range values {
			sb.WriteString("&")
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
		}
	}
	if len(body) > 0 {
		sb.WriteString("&http_body=")
		sb.Write(body)
	}
	return sb.String()
}

// ComputeNewSignature 新版签名：SHA-256 小写 hex（放 header x-life-sign）
func ComputeNewSignature(str1 string) string {
	sum := sha256.Sum256([]byte(str1))
	return hex.EncodeToString(sum[:])
}

// ComputeOldSignature 旧版签名：MD5 小写 hex（放 URL 参数 sign）
func ComputeOldSignature(str1 string) string {
	sum := md5.Sum([]byte(str1))
	return hex.EncodeToString(sum[:])
}

// VerifySPIRequest 校验抖音 SPI 回调请求的签名。
// 新版取 header x-life-sign 做 SHA-256 校验，否则取 URL 参数 sign 做 MD5 校验，
// 两者都没有直接判失败。校验通过后会把已读取的 body 重新写回 r.Body，供后续业务绑定使用。
func VerifySPIRequest(r *http.Request, clientSecret string) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.Join(err, fmt.Errorf("douyin SPI 读取 body 失败"))
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	str1 := BuildSignString(clientSecret, r.URL.Query(), body)
	if newSign := r.Header.Get("x-life-sign"); newSign != "" {
		if strings.EqualFold(newSign, ComputeNewSignature(str1)) {
			return nil
		}
		return fmt.Errorf("douyin SPI 新版签名(x-life-sign)校验失败")
	}
	if oldSign := r.URL.Query().Get("sign"); oldSign != "" {
		if strings.EqualFold(oldSign, ComputeOldSignature(str1)) {
			return nil
		}
		return fmt.Errorf("douyin SPI 旧版签名(sign)校验失败")
	}
	return fmt.Errorf("douyin SPI 请求缺少签名(x-life-sign / sign 均为空)")
}

// SpiSignVerify gin 中间件：校验 SPI 回调签名 + client_key。
// 验签失败按抖音 SPI 出参格式返回 error_code=3000007(操作无权限) 并 Abort。
func SpiSignVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientSecret := variable.ConfigYml.GetString("Douyin.ClientSecret")
		if clientSecret == "" {
			variable.ZapLog.Error("douyin SPI 验签失败：配置 Douyin.ClientSecret 为空")
			c.JSON(http.StatusOK, gin.H{"data": gin.H{"error_code": 3000007, "description": "sign verify failed"}})
			c.Abort()
			return
		}
		// URL 固定携带 client_key，与配置不一致说明不是发给我们这个应用的回调
		if ck := c.Query("client_key"); ck != "" {
			if want := variable.ConfigYml.GetString("Douyin.ClientKey"); want != "" && ck != want {
				variable.ZapLog.Warn("douyin SPI client_key 不匹配", zap.String("got", ck))
				c.JSON(http.StatusOK, gin.H{"data": gin.H{"error_code": 3000007, "description": "client_key mismatch"}})
				c.Abort()
				return
			}
		}

		if err := VerifySPIRequest(c.Request, clientSecret); err != nil {
			variable.ZapLog.Warn("douyin SPI 验签失败",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
				zap.String("logid", c.GetHeader("X-Bytedance-Logid")))
			c.JSON(http.StatusOK, gin.H{"data": gin.H{"error_code": 3000007, "description": "sign verify failed"}})
			c.Abort()
			return
		}
		c.Next()
	}
}
