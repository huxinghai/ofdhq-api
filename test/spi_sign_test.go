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
	"ofdhq-api/pkg/verify"
)

// SPI 验签纯函数单测（不触网、不依赖数据库）。
// 签名算法在 pkg/verify 包实现（API 与官方 douyin-sign-verify-go/pkg/verify 示例一致）。
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
	got := verify.BuildSignString(spiTestClientSecret, q, body)
	// 注意：secret 与各参数之间均为单个 &
	want := spiTestClientSecret +
		"&client_key=ck123" +
		"&other=a&other=b" +
		"&timestamp=1710000000000" +
		`&http_body={"order_id":"1"}`
	if got != want {
		t.Fatalf("BuildSignString 拼装结果不符:\ngot:  %s\nwant: %s", got, want)
	}

	// 空 body 不拼 http_body 段
	if noBody := verify.BuildSignString(spiTestClientSecret, q, nil); bytes.Contains([]byte(noBody), []byte("http_body")) {
		t.Fatalf("空 body 不应拼接 http_body, got: %s", noBody)
	}
}

func TestDouyinComputeSignatures(t *testing.T) {
	q := url.Values{}
	q.Set("client_key", "ck")
	body := []byte(`{}`)
	str1 := verify.BuildSignString("secret", q, body)
	// 与标准库独立计算结果比对，保证是小写 hex
	shaSum := sha256.Sum256([]byte(str1))
	if got := verify.ComputeNewSignature("secret", q, body); got != hex.EncodeToString(shaSum[:]) {
		t.Fatalf("ComputeNewSignature 不匹配: %s", got)
	}
	md5Sum := md5.Sum([]byte(str1))
	if got := verify.ComputeOldSignature("secret", q, body); got != hex.EncodeToString(md5Sum[:]) {
		t.Fatalf("ComputeOldSignature 不匹配: %s", got)
	}
}

func TestDouyinVerifySPIRequestNewSign(t *testing.T) {
	body := []byte(`{"order_id":"o1","pay_amount":100}`)
	q := url.Values{}
	q.Set("client_key", "ck123")
	q.Set("timestamp", "1710000000000")

	// 新版：header x-life-sign (SHA-256)
	req := httptest.NewRequest("POST", "/callback?"+q.Encode(), bytes.NewReader(body))
	req.Header.Set("x-life-sign", verify.ComputeNewSignature(spiTestClientSecret, q, body))
	if err := douyin.VerifySPIRequest(req, spiTestClientSecret); err != nil {
		t.Fatalf("新版签名校验应通过: %v", err)
	}
	// 校验后 body 已还原，可再次读取（下游控制器依赖这一点绑定参数）
	if again, _ := io.ReadAll(req.Body); !bytes.Equal(again, body) {
		t.Fatalf("验签后 body 未正确还原: %s", again)
	}

	// 篡改 body 必须失败
	req2 := httptest.NewRequest("POST", "/callback?"+q.Encode(), bytes.NewReader([]byte(`{"order_id":"o2"}`)))
	req2.Header.Set("x-life-sign", verify.ComputeNewSignature(spiTestClientSecret, q, body))
	if err := douyin.VerifySPIRequest(req2, spiTestClientSecret); err == nil {
		t.Fatal("篡改 body 后新版签名校验应失败")
	}

	// 篡改 query 必须失败
	q2 := url.Values{}
	q2.Set("client_key", "ck999")
	q2.Set("timestamp", "1710000000000")
	req3 := httptest.NewRequest("POST", "/callback?"+q2.Encode(), bytes.NewReader(body))
	req3.Header.Set("x-life-sign", verify.ComputeNewSignature(spiTestClientSecret, q, body))
	if err := douyin.VerifySPIRequest(req3, spiTestClientSecret); err == nil {
		t.Fatal("篡改 query 后新版签名校验应失败")
	}
}

func TestDouyinVerifySPIRequestOldSign(t *testing.T) {
	body := []byte(`{"order_id":"o1"}`)
	q := url.Values{}
	q.Set("client_key", "ck123")
	q.Set("timestamp", "1710000000000")

	q.Set("sign", verify.ComputeOldSignature(spiTestClientSecret, q, body)) // sign 参数自身不参与签名

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

// ---- 线上真实回调回归用例（2026-09 生产抓包，修复 secret 后双 & 拼错问题时锁定） ----
// 抖音真实回调对同一个 str1 同时携带：旧版 URL 参数 sign(MD5) + 新版 header x-life-sign(SHA-256)。
// 本用例用真实密钥与真实报文锁定 str1 布局（secret 后单个 &），防止拼装再次回归；
// 同时按官方 demo 的用法直接调用 pkg/verify 做 SDK 式校验。
func TestDouyinVerifySPIRequestRealCallback(t *testing.T) {
	const realSecret = "30d07a13646ecb6d5b5604fa6d1203ee"
	const realOldSign = "9699a38f36a590a0639c837a98ddde5a"                               // URL 参数 sign
	const realNewSign = "8cdea8c856bb41e1dcb3b8bebe14690a05058282d98f3cc13ca4382595b01c50" // header x-life-sign
	body := []byte(`{"create_order_time_unix":1790574818,"total_coupon_count":1,"pay_amount":1,"commerce_info":{},"departure_info":null,"each_coupon_amount":1,"actual_amount":1,"merchant_discount_amount":0,"total_amount":1,"discount_amount":0,"buyer_info":{"phone":"CjcCD96ugJqF9rARBFKNaw=="},"product_snap_shot":[{"commodity":[{"group_name":"景点门票","total_count":1,"option_count":1,"item_list":[{"name":"球球","price":2,"count":1,"unit":"份"}]}],"earliest_appointment":{"need_appointment":true,"ahead_time_type":1,"ahead_time":10,"change_time_type":0},"ticket_type":"BigTicket","product_id":"1877536534475792","create_time":1790560038,"add_price_policy":{"enable":false},"appointment":{"need_appointment":true,"ahead_time_type":1,"ahead_time":1,"change_time_type":0},"refund_rule":{"refund_type":2,"calc_price_type":0,"refund_rule_type":0},"use_location":{"custom_location":"","scenic_location_flag":true},"sku_id":"1877536534475792","out_product_id":"13702234|658692|3641435-7324904536798152719","update_time":1790574622,"is_superimposed_discounts":false,"use_date":{"day_duration":360,"use_date_type":2},"self_funded_project":{"enable":false},"name":"【测试勿购灰度景区kaka】","category_id":"18004001","category_full_name":"游玩·景点票券·单景点门票","description_rich_text":[{"note_type":1,"content":""}],"region":{"restrict_flag":false},"IsNeedPick":false,"use_time":{"use_time_type":1}}],"item_list":[{"item_id":"800000586844010243713853719"}],"order_id":"1113338212130293719","biz_type":3011,"pay_info":{"order_source":1,"pay_time_unix":1790574824}}`)

	q := url.Values{}
	q.Set("client_key", "aw71tlvaotendxw0")
	q.Set("timestamp", "1790574887543")
	q.Set("sign", realOldSign)

	// 官方 demo 用法：VerifyNewSignatureFromRequest / VerifyOldSignatureFromRequest
	reqSdk := httptest.NewRequest("POST", "/api/v1/douyin/spi/presale_order/create?"+q.Encode(), bytes.NewReader(body))
	reqSdk.Header.Set("x-life-sign", realNewSign)
	if ok, expected, provided := verify.VerifyNewSignatureFromRequest(reqSdk, realSecret); !ok {
		t.Fatalf("SDK 式新版签名校验应通过: expected=%s provided=%s", expected, provided)
	} else if expected != realNewSign || provided != realNewSign {
		t.Fatalf("新版签名值不符: expected=%s provided=%s", expected, provided)
	}
	reqSdkOld := httptest.NewRequest("POST", "/api/v1/douyin/spi/presale_order/create?"+q.Encode(), bytes.NewReader(body))
	if ok, expected, provided := verify.VerifyOldSignatureFromRequest(reqSdkOld, realSecret); !ok {
		t.Fatalf("SDK 式旧版签名校验应通过: expected=%s provided=%s", expected, provided)
	}

	// 中间件入口 VerifySPIRequest：新旧签名同时携带（抖音线上实际行为），优先走新版校验
	req2 := httptest.NewRequest("POST", "/api/v1/douyin/spi/presale_order/create?"+q.Encode(), bytes.NewReader(body))
	req2.Header.Set("x-life-sign", realNewSign)
	req2.Header.Set("x-life-clientkey", "aw71tlvaotendxw0")
	if err := douyin.VerifySPIRequest(req2, realSecret); err != nil {
		t.Fatalf("线上真实回调新版签名(x-life-sign)校验应通过: %v", err)
	}

	// 仅旧版 URL sign（无 x-life-sign header）也应通过
	req3 := httptest.NewRequest("POST", "/api/v1/douyin/spi/presale_order/create?"+q.Encode(), bytes.NewReader(body))
	if err := douyin.VerifySPIRequest(req3, realSecret); err != nil {
		t.Fatalf("线上真实回调旧版签名(sign)校验应通过: %v", err)
	}
}
