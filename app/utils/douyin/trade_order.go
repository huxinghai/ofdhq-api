package douyin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"ofdhq-api/app/global/variable"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/bytedance/douyin-openapi-sdk-go/client"
)

// goodlife 交易订单查询端点。不使用 SDK 的 cli.TradeOrderQuery：
// SDK 会把未设置的字段以零值拼进 query（nil *int64 经 tea.Int64Value 变成 0，
// StringifyMapValue 只过滤 nil），导致 URL 固定带 update_order_start_time=0&
// update_order_end_time=0，与 create_order_* 时间窗冲突时服务端查不到数据。
const tradeOrderQueryURL = "https://open.douyin.com/goodlife/v1/trade/order/query/"

// 注意：goodlife 大时间范围查询响应可能较慢，超时给足；DNS 被 Fake-IP（代理 TUN 模式）劫持、
// 国内域名走国外节点时会出现 awaiting headers 超时，需在代理侧给 douyin.com 配直连规则
const httpTimeout = 30 * time.Second

var httpClient = &http.Client{Timeout: httpTimeout}

// TradeOrderQueryParam 交易订单查询参数，只发送显式设置（非零值）的字段
type TradeOrderQueryParam struct {
	AccountID            string // 为空时取配置 Douyin.AccountId
	CreateOrderStartTime int64  // 下单开始时间（秒级时间戳），可选
	CreateOrderEndTime   int64  // 下单结束时间（秒级时间戳），可选
	OrderID              string // 订单号，可选
	OrderStatus          *int32 // 订单状态，可选
	PageNum              int32  // 页码，默认 1
	PageSize             int32  // 每页数量，默认 10
}

// TradeOrderQuery 查询本地生活（goodlife）交易订单，分页数据见 resp.Data.Orders / resp.Data.Page
func TradeOrderQuery(p *TradeOrderQueryParam) (*client.TradeOrderQueryResponse, error) {
	if p == nil {
		return nil, fmt.Errorf("douyin TradeOrderQuery param is nil")
	}
	accountID := p.AccountID
	if accountID == "" {
		accountID = variable.ConfigYml.GetString("Douyin.AccountId")
	}
	if accountID == "" {
		return nil, fmt.Errorf("douyin TradeOrderQuery 缺少 account_id（参数与配置 Douyin.AccountId 均为空）")
	}
	pageNum, pageSize := p.PageNum, p.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	accessToken, err := GetAccessToken()
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("account_id", accountID)
	if p.CreateOrderStartTime > 0 {
		q.Set("create_order_start_time", strconv.FormatInt(p.CreateOrderStartTime, 10))
	}
	if p.CreateOrderEndTime > 0 {
		q.Set("create_order_end_time", strconv.FormatInt(p.CreateOrderEndTime, 10))
	}
	if p.OrderID != "" {
		q.Set("order_id", p.OrderID)
	}
	if p.OrderStatus != nil {
		q.Set("order_status", strconv.FormatInt(int64(*p.OrderStatus), 10))
	}
	q.Set("page_num", strconv.FormatInt(int64(pageNum), 10))
	q.Set("page_size", strconv.FormatInt(int64(pageSize), 10))

	req, err := http.NewRequest(http.MethodGet, tradeOrderQueryURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("douyin TradeOrderQuery NewRequest"))
	}
	req.Header.Set("access-token", accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("douyin TradeOrderQuery 请求失败"))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("douyin TradeOrderQuery 读取响应失败"))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("douyin TradeOrderQuery HTTP %d, body=%s", resp.StatusCode, body)
	}

	var result client.TradeOrderQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Join(err, fmt.Errorf("douyin TradeOrderQuery 解析响应失败, body=%s", body))
	}
	if result.Extra != nil && result.Extra.ErrorCode != nil && *result.Extra.ErrorCode != 0 {
		return &result, fmt.Errorf("douyin TradeOrderQuery 失败: error_code=%d description=%s logid=%s",
			*result.Extra.ErrorCode, tea.StringValue(result.Extra.Description), tea.StringValue(result.Extra.Logid))
	}
	return &result, nil
}
