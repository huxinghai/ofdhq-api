// Package verify 实现抖音本地生活 SPI 回调的签名计算与校验。
//
// API 与官方文档示例 SDK `douyin-sign-verify-go/pkg/verify` 保持一致，
// 该 SDK 未在 GitHub/Gitee 公开发布，本包按官方签名规则文档实现：
// https://partner.open-douyin.com/docs/resource/zh-CN/local-life/develop/preparation/spi-signature-rules
//
// 待签名明文 str1 的拼装规则（各部分以单个 & 连接，已用线上真实回调比对验证）：
//  1. 第一部分为 client_secret
//  2. 取 URL 上除 sign（大小写不敏感）以外的全部 query 参数，按 key 字典升序排列；
//     同一个 key 有多个 value 时 value 也升序排列，每部分形如 "key=value"
//  3. body 不为空时，末尾追加 "http_body=" + 原始 body 字符串
//     即 str1 = client_secret&k1=v1&k2=v2&...&http_body=<body>
//     注意 secret 后面是单个 &，不要多拼（否则与抖音侧不一致导致验签失败）
//
// 新版签名：header x-life-sign = SHA-256(str1) 的小写 hex
// 旧版签名：URL 参数 sign    = MD5(str1)     的小写 hex
//
// 注意：签名针对的是抖音发来的原始 body 字节，必须先读原始 body 再验签，
// 严禁先 json.Unmarshal 再重新序列化（字段顺序/转义变化会导致验签失败）。
// 两个 Verify*FromRequest 读取 body 后都会还原 r.Body，可重复读取。
package verify

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// BuildSignString 按官方规则拼接待签名明文 str1（导出便于联调排查）
func BuildSignString(clientSecret string, query url.Values, body []byte) string {
	var sb strings.Builder
	sb.WriteString(clientSecret)

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

// ComputeNewSignature 新版签名：SHA-256 小写 hex（放 header x-life-sign），
// 抖音侧调用我们的 SPI 前用它计算签名；也可用于本地模拟抖音侧构造请求。
func ComputeNewSignature(clientSecret string, query url.Values, body []byte) string {
	sum := sha256.Sum256([]byte(BuildSignString(clientSecret, query, body)))
	return hex.EncodeToString(sum[:])
}

// ComputeOldSignature 旧版签名：MD5 小写 hex（放 URL 参数 sign）
func ComputeOldSignature(clientSecret string, query url.Values, body []byte) string {
	sum := md5.Sum([]byte(BuildSignString(clientSecret, query, body)))
	return hex.EncodeToString(sum[:])
}

// readAndRestoreBody 读取请求原始 body 并还原，供重复读取
func readAndRestoreBody(r *http.Request) ([]byte, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, false
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, true
}

// VerifyNewSignatureFromRequest 从请求 header x-life-sign 校验新版签名。
// 返回 (是否通过, 我方计算的期望签名, 请求携带的签名)。
func VerifyNewSignatureFromRequest(r *http.Request, clientSecret string) (ok bool, expected, provided string) {
	body, readable := readAndRestoreBody(r)
	if !readable {
		return false, "", ""
	}
	expected = ComputeNewSignature(clientSecret, r.URL.Query(), body)
	provided = r.Header.Get("x-life-sign")
	if provided == "" {
		return false, expected, provided
	}
	return strings.EqualFold(expected, provided), expected, provided
}

// VerifyOldSignatureFromRequest 从 URL 参数 sign 校验旧版签名。
// 返回 (是否通过, 我方计算的期望签名, 请求携带的签名)。
func VerifyOldSignatureFromRequest(r *http.Request, clientSecret string) (ok bool, expected, provided string) {
	body, readable := readAndRestoreBody(r)
	if !readable {
		return false, "", ""
	}
	expected = ComputeOldSignature(clientSecret, r.URL.Query(), body)
	provided = r.URL.Query().Get("sign")
	if provided == "" {
		return false, expected, provided
	}
	return strings.EqualFold(expected, provided), expected, provided
}
