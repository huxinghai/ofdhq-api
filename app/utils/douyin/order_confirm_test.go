package douyin

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TravelAgencyOrderConfirm 单测（包内测试，便于替换未导出的注入点：
// travelAgencyOrderConfirmURL / getAccessToken），全部走本地 httptest 服务，不触网。

// stubConfirmServer 把接口地址替换为本地 httptest 服务（保留真实路径便于断言），
// token 固定返回，测试结束自动恢复现场。
func stubConfirmServer(t *testing.T, token string, tokenErr error, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	origURL, origToken := travelAgencyOrderConfirmURL, getAccessToken
	travelAgencyOrderConfirmURL = srv.URL + "/goodlife/v1/trip/trade/travelagency/order/confirm/"
	getAccessToken = func() (string, error) { return token, tokenErr }
	t.Cleanup(func() {
		srv.Close()
		travelAgencyOrderConfirmURL, getAccessToken = origURL, origToken
	})
}

func TestTravelAgencyOrderConfirmParamInvalid(t *testing.T) {
	// 参数校验发生在取 token 之前，无需桩服务即可断言
	cases := []struct {
		name string
		p    *TravelAgencyOrderConfirmParam
		want string
	}{
		{"nil 参数", nil, "param is nil"},
		{"缺 order_id", &TravelAgencyOrderConfirmParam{SourceOrderID: "s1", ConfirmResult: ConfirmResultAccept}, "order_id"},
		{"缺 source_order_id", &TravelAgencyOrderConfirmParam{OrderID: "o1", ConfirmResult: ConfirmResultAccept}, "source_order_id"},
		{"confirm_result 非法0", &TravelAgencyOrderConfirmParam{OrderID: "o1", SourceOrderID: "s1"}, "confirm_result 非法"},
		{"confirm_result 非法3", &TravelAgencyOrderConfirmParam{OrderID: "o1", SourceOrderID: "s1", ConfirmResult: 3}, "confirm_result 非法"},
		{"拒单缺 reject_code", &TravelAgencyOrderConfirmParam{OrderID: "o1", SourceOrderID: "s1", ConfirmResult: ConfirmResultReject}, "reject_code"},
	}
	for _, c := range cases {
		if err := TravelAgencyOrderConfirm(c.p); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("[%s] 期望报错含 %q, got: %v", c.name, c.want, err)
		}
	}
}

func TestTravelAgencyOrderConfirmAccept(t *testing.T) {
	var gotReq struct {
		Method string
		Path   string
		Token  string
		CType  string
		Body   map[string]any
	}
	stubConfirmServer(t, "test-token", nil, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		gotReq.Method, gotReq.Path = r.Method, r.URL.Path
		gotReq.Token, gotReq.CType = r.Header.Get("access-token"), r.Header.Get("content-type")
		_ = json.Unmarshal(raw, &gotReq.Body)
		w.Write([]byte(`{"data":{"error_code":0,"description":"success","order_id":"o1","order_out_id":"out1"},"extra":{"error_code":0,"description":"success","logid":"lg1"}}`))
	})

	if err := TravelAgencyOrderConfirm(&TravelAgencyOrderConfirmParam{
		OrderID: "o1", SourceOrderID: "s1", ConfirmResult: ConfirmResultAccept,
	}); err != nil {
		t.Fatalf("接单应成功: %v", err)
	}

	// 请求侧断言：方法/路径/鉴权头/报文结构
	if gotReq.Method != http.MethodPost {
		t.Errorf("method 应为 POST, got %s", gotReq.Method)
	}
	if gotReq.Path != "/goodlife/v1/trip/trade/travelagency/order/confirm/" {
		t.Errorf("path 不符: %s", gotReq.Path)
	}
	if gotReq.Token != "test-token" {
		t.Errorf("access-token 头应携带注入的 token, got %q", gotReq.Token)
	}
	if gotReq.CType != "application/json" {
		t.Errorf("content-type 应为 application/json, got %q", gotReq.CType)
	}
	if gotReq.Body["order_id"] != "o1" || gotReq.Body["source_order_id"] != "s1" {
		t.Errorf("order_id/source_order_id 不符: %v", gotReq.Body)
	}
	ci, _ := gotReq.Body["confirm_info"].(map[string]any)
	if ci == nil || ci["confirm_result"] != float64(ConfirmResultAccept) {
		t.Errorf("confirm_info.confirm_result 应为接单(1): %v", gotReq.Body)
	}
	// 接单时 omitempty：不应出现 reject_code / extra_msg
	if _, ok := ci["reject_code"]; ok {
		t.Errorf("接单报文不应携带 reject_code: %v", gotReq.Body)
	}
	if _, ok := ci["extra_msg"]; ok {
		t.Errorf("接单报文不应携带 extra_msg: %v", gotReq.Body)
	}
}

func TestTravelAgencyOrderConfirmReject(t *testing.T) {
	var body map[string]any
	stubConfirmServer(t, "test-token", nil, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Write([]byte(`{"data":{"error_code":0,"description":"success"}}`))
	})

	if err := TravelAgencyOrderConfirm(&TravelAgencyOrderConfirmParam{
		OrderID: "o1", SourceOrderID: "s1",
		ConfirmResult: ConfirmResultReject, RejectCode: RejectCodeSoldOut, ExtraMsg: "当日库存已约满",
	}); err != nil {
		t.Fatalf("拒单应成功: %v", err)
	}

	ci, _ := body["confirm_info"].(map[string]any)
	if ci == nil || ci["confirm_result"] != float64(ConfirmResultReject) {
		t.Fatalf("confirm_result 应为拒单(2): %v", body)
	}
	if ci["reject_code"] != float64(RejectCodeSoldOut) {
		t.Errorf("reject_code 应为 1(库存已约满): %v", body)
	}
	if ci["extra_msg"] != "当日库存已约满" {
		t.Errorf("extra_msg 不符: %v", body)
	}
}

func TestTravelAgencyOrderConfirmBizError(t *testing.T) {
	stubConfirmServer(t, "test-token", nil, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"error_code":2190018,"description":"订单状态不允许确认"}}`))
	})
	err := TravelAgencyOrderConfirm(&TravelAgencyOrderConfirmParam{
		OrderID: "o1", SourceOrderID: "s1", ConfirmResult: ConfirmResultAccept,
	})
	if err == nil || !strings.Contains(err.Error(), "2190018") || !strings.Contains(err.Error(), "订单状态不允许确认") {
		t.Fatalf("应透出 data 层业务错误, got: %v", err)
	}
}

func TestTravelAgencyOrderConfirmExtraError(t *testing.T) {
	stubConfirmServer(t, "test-token", nil, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"error_code":0},"extra":{"error_code":1,"description":"param error","logid":"log-123"}}`))
	})
	err := TravelAgencyOrderConfirm(&TravelAgencyOrderConfirmParam{
		OrderID: "o1", SourceOrderID: "s1", ConfirmResult: ConfirmResultAccept,
	})
	if err == nil || !strings.Contains(err.Error(), "log-123") {
		t.Fatalf("应透出 extra 层错误(含 logid), got: %v", err)
	}
}

func TestTravelAgencyOrderConfirmHTTPError(t *testing.T) {
	stubConfirmServer(t, "test-token", nil, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	})
	err := TravelAgencyOrderConfirm(&TravelAgencyOrderConfirmParam{
		OrderID: "o1", SourceOrderID: "s1", ConfirmResult: ConfirmResultAccept,
	})
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("应透出 HTTP 状态码错误, got: %v", err)
	}
}

func TestTravelAgencyOrderConfirmBadJSON(t *testing.T) {
	stubConfirmServer(t, "test-token", nil, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	})
	err := TravelAgencyOrderConfirm(&TravelAgencyOrderConfirmParam{
		OrderID: "o1", SourceOrderID: "s1", ConfirmResult: ConfirmResultAccept,
	})
	if err == nil || !strings.Contains(err.Error(), "解析响应失败") {
		t.Fatalf("应报响应解析失败, got: %v", err)
	}
}

func TestTravelAgencyOrderConfirmTokenFail(t *testing.T) {
	called := 0
	stubConfirmServer(t, "", fmtTokenErr, func(w http.ResponseWriter, r *http.Request) {
		called++
		w.Write([]byte(`{"data":{"error_code":0}}`))
	})
	err := TravelAgencyOrderConfirm(&TravelAgencyOrderConfirmParam{
		OrderID: "o1", SourceOrderID: "s1", ConfirmResult: ConfirmResultAccept,
	})
	if err == nil || !strings.Contains(err.Error(), "redis down") {
		t.Fatalf("应透出取 token 失败错误, got: %v", err)
	}
	if called != 0 {
		t.Fatalf("取 token 失败不应发起 HTTP 请求, 实际请求 %d 次", called)
	}
}

var fmtTokenErr = &simpleErr{"redis down"}

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }

func TestTravelAgencyOrderConfirmNetworkError(t *testing.T) {
	// 先起服务拿到地址再关掉，模拟连接失败
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL + "/goodlife/v1/trip/trade/travelagency/order/confirm/"
	srv.Close()

	origURL, origToken := travelAgencyOrderConfirmURL, getAccessToken
	travelAgencyOrderConfirmURL = url
	getAccessToken = func() (string, error) { return "test-token", nil }
	defer func() { travelAgencyOrderConfirmURL, getAccessToken = origURL, origToken }()

	err := TravelAgencyOrderConfirm(&TravelAgencyOrderConfirmParam{
		OrderID: "o1", SourceOrderID: "s1", ConfirmResult: ConfirmResultAccept,
	})
	if err == nil || !strings.Contains(err.Error(), "请求失败") {
		t.Fatalf("应报网络请求失败, got: %v", err)
	}
}

func TestTravOrderConfirmBizError(t *testing.T) {
	err := TravelAgencyOrderConfirm(&TravelAgencyOrderConfirmParam{
		OrderID: "o1", SourceOrderID: "s1", ConfirmResult: ConfirmResultAccept,
	})
	if err == nil || !strings.Contains(err.Error(), "2190018") || !strings.Contains(err.Error(), "订单状态不允许确认") {
		t.Fatalf("应透出 data 层业务错误, got: %v", err)
	}
}
