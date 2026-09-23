package test

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http/httptest"
	"net/url"
	"testing"

	douyin "ofdhq-api/app/utils/douyin"
)

// SPI 验签纯函数单测（不触网、不依赖数据库）。
// 规则文档：https://partner.open-douyin.com/docs/resource/zh-CN/local-life/develop/preparation/spi-signature-rules
const spiTestClientSecret = "test_secret"

func TestDouyinBuildSignString(t *testing.T) {
	// key 字典升序 + 同 key 多 value 升序 + sign 不参与签名 + body 拼在末尾
	q := url.Values{}
	q.Set("client_key", "ck123")
	q.Set("timestamp", "1710000000000")
	q.Add("other", "b")
	q.Add("other", "a")
	q.Set("sign", "should_be_excluded")

	body := []byte(`{"order_id":"1"}`)
	got := douyin.BuildSignString(spiTestClientSecret, q, body)
	want := spiTestClientSecret + "&" +
		"&client_key=ck123" +
		"&other=a&other=b" +
		"&timestamp=1710000000000" +
		`&http_body={"order_id":"1"}`
	if got != want {
		t.Fatalf("BuildSignString 拼装结果不符:\ngot:  %s\nwant: %s", got, want)
	}

	// 空 body 不拼 http_body 段
	if noBody := douyin.BuildSignString(spiTestClientSecret, q, nil); bytes.Contains([]byte(noBody), []byte("http_body")) {
		t.Fatalf("空 body 不应拼接 http_body, got: %s", noBody)
	}
}

func TestDouyinComputeSignatures(t *testing.T) {
	str1 := "secret&client_key=ck&http_body={}"
	// 与标准库独立计算结果比对，保证是小写 hex
	shaSum := sha256.Sum256([]byte(str1))
	if got := douyin.ComputeNewSignature(str1); got != hex.EncodeToString(shaSum[:]) {
		t.Fatalf("ComputeNewSignature 不匹配: %s", got)
	}
	md5Sum := md5.Sum([]byte(str1))
	if got := douyin.ComputeOldSignature(str1); got != hex.EncodeToString(md5Sum[:]) {
		t.Fatalf("ComputeOldSignature 不匹配: %s", got)
	}
}

func TestDouyinVerifySPIRequestNewSign(t *testing.T) {
	body := []byte(`{"order_id":"o1","pay_amount":100}`)
	q := url.Values{}
	q.Set("client_key", "ck123")
	q.Set("timestamp", "1710000000000")

	str1 := douyin.BuildSignString(spiTestClientSecret, q, body)

	// 新版：header x-life-sign (SHA-256)
	req := httptest.NewRequest("POST", "/callback?"+q.Encode(), bytes.NewReader(body))
	req.Header.Set("x-life-sign", douyin.ComputeNewSignature(str1))
	if err := douyin.VerifySPIRequest(req, spiTestClientSecret); err != nil {
		t.Fatalf("新版签名校验应通过: %v", err)
	}
	// 校验后 body 已还原，可再次读取（下游控制器依赖这一点绑定参数）
	if again, _ := io.ReadAll(req.Body); !bytes.Equal(again, body) {
		t.Fatalf("验签后 body 未正确还原: %s", again)
	}

	// 篡改 body 必须失败
	req2 := httptest.NewRequest("POST", "/callback?"+q.Encode(), bytes.NewReader([]byte(`{"order_id":"o2"}`)))
	req2.Header.Set("x-life-sign", douyin.ComputeNewSignature(str1))
	if err := douyin.VerifySPIRequest(req2, spiTestClientSecret); err == nil {
		t.Fatal("篡改 body 后新版签名校验应失败")
	}

	// 篡改 query 必须失败
	q2 := url.Values{}
	q2.Set("client_key", "ck999")
	q2.Set("timestamp", "1710000000000")
	req3 := httptest.NewRequest("POST", "/callback?"+q2.Encode(), bytes.NewReader(body))
	req3.Header.Set("x-life-sign", douyin.ComputeNewSignature(str1))
	if err := douyin.VerifySPIRequest(req3, spiTestClientSecret); err == nil {
		t.Fatal("篡改 query 后新版签名校验应失败")
	}
}

func TestDouyinVerifySPIRequestOldSign(t *testing.T) {
	body := []byte(`{"order_id":"o1"}`)
	q := url.Values{}
	q.Set("client_key", "ck123")
	q.Set("timestamp", "1710000000000")

	str1 := douyin.BuildSignString(spiTestClientSecret, q, body)
	q.Set("sign", douyin.ComputeOldSignature(str1)) // sign 参数自身不参与签名

	req := httptest.NewRequest("POST", "/callback?"+q.Encode(), bytes.NewReader(body))
	if err := douyin.VerifySPIRequest(req, spiTestClientSecret); err != nil {
		t.Fatalf("旧版签名校验应通过: %v", err)
	}
}

func TestDouyinVerifySPIRequestNoSign(t *testing.T) {
	req := httptest.NewRequest("POST", "/callback?client_key=ck123", bytes.NewReader([]byte(`{}`)))
	if err := douyin.VerifySPIRequest(req, spiTestClientSecret); err == nil {
		t.Fatal("缺少签名（x-life-sign / sign 均为空）应校验失败")
	}
}
