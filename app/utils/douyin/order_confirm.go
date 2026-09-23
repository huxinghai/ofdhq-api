package douyin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alibabacloud-go/tea/tea"
)

// 旅行社交易确认接单接口（我方 -> 抖音，OpenAPI）
// POST /goodlife/v1/trip/trade/travelagency/order/confirm/
// 抖音侧通知支付成功后，需在商品设置的接单时间内异步确认（超时会拒单），
// 官方建议对确认接口异步重试 3 次、间隔 5s。
const travelAgencyOrderConfirmURL = "https://open.douyin.com/goodlife/v1/trip/trade/travelagency/order/confirm/"

// 确认结果
const (
	ConfirmResultAccept int32 = 1 // 接单
	ConfirmResultReject int32 = 2 // 拒单
)

// 拒单码（confirm_result=2 时使用）
const (
	RejectCodeSoldOut      int32 = 1 // 库存已约满
	RejectCodeNeedAddPrice int32 = 2 // 商品需加价
	RejectCodeCannotMeet   int32 = 3 // 无法满足顾客需求
)

// TravelAgencyOrderConfirmParam 确认接单参数（只实现必填 + 拒单所需字段）
type TravelAgencyOrderConfirmParam struct {
	OrderID       string // 抖音侧预约订单号
	SourceOrderID string // 预约订单归属的预售订单ID
	ConfirmResult int32  // 1接单 2拒单
	RejectCode    int32  // 拒单时必填：1库存已约满 2商品需加价 3无法满足顾客需求
	ExtraMsg      string // 拒单原因等补充说明，纯文本 <=200 字
}

type travelAgencyOrderConfirmReq struct {
	OrderID       string `json:"order_id"`
	SourceOrderID string `json:"source_order_id"`
	ConfirmInfo   struct {
		ConfirmResult int32  `json:"confirm_result"`
		RejectCode    int32  `json:"reject_code,omitempty"`
		ExtraMsg      string `json:"extra_msg,omitempty"`
	} `json:"confirm_info"`
}

type travelAgencyOrderConfirmResp struct {
	Data *struct {
		ErrorCode  *int32  `json:"error_code"`
		Description *string `json:"description"`
		OrderID    *string `json:"order_id"`
		OrderOutID *string `json:"order_out_id"`
	} `json:"data"`
	Extra *struct {
		ErrorCode   *int32  `json:"error_code"`
		Description *string `json:"description"`
		Logid       *string `json:"logid"`
	} `json:"extra"`
}

// TravelAgencyOrderConfirm 通知抖音接单/拒单结果。
// 只校验接口侧必填项（order_id、source_order_id、confirm_result；拒单时 reject_code），
// 预定信息（free_travel_info/hotel_info/play_info）为可选字段，暂不填写。
func TravelAgencyOrderConfirm(p *TravelAgencyOrderConfirmParam) error {
	if p == nil {
		return fmt.Errorf("douyin TravelAgencyOrderConfirm param is nil")
	}
	if p.OrderID == "" || p.SourceOrderID == "" {
		return fmt.Errorf("douyin TravelAgencyOrderConfirm 缺少 order_id / source_order_id")
	}
	if p.ConfirmResult != ConfirmResultAccept && p.ConfirmResult != ConfirmResultReject {
		return fmt.Errorf("douyin TravelAgencyOrderConfirm confirm_result 非法: %d", p.ConfirmResult)
	}
	if p.ConfirmResult == ConfirmResultReject && p.RejectCode == 0 {
		return fmt.Errorf("douyin TravelAgencyOrderConfirm 拒单时必须携带 reject_code")
	}

	reqBody := travelAgencyOrderConfirmReq{
		OrderID:       p.OrderID,
		SourceOrderID: p.SourceOrderID,
	}
	reqBody.ConfirmInfo.ConfirmResult = p.ConfirmResult
	reqBody.ConfirmInfo.RejectCode = p.RejectCode
	reqBody.ConfirmInfo.ExtraMsg = p.ExtraMsg

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return errors.Join(err, fmt.Errorf("douyin TravelAgencyOrderConfirm 序列化请求失败"))
	}

	accessToken, err := GetAccessToken()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, travelAgencyOrderConfirmURL, bytes.NewReader(payload))
	if err != nil {
		return errors.Join(err, fmt.Errorf("douyin TravelAgencyOrderConfirm NewRequest"))
	}
	req.Header.Set("access-token", accessToken)
	req.Header.Set("content-type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Join(err, fmt.Errorf("douyin TravelAgencyOrderConfirm 请求失败"))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Join(err, fmt.Errorf("douyin TravelAgencyOrderConfirm 读取响应失败"))
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("douyin TravelAgencyOrderConfirm HTTP %d, body=%s", resp.StatusCode, body)
	}

	var result travelAgencyOrderConfirmResp
	if err := json.Unmarshal(body, &result); err != nil {
		return errors.Join(err, fmt.Errorf("douyin TravelAgencyOrderConfirm 解析响应失败, body=%s", body))
	}
	if result.Data != nil && result.Data.ErrorCode != nil && *result.Data.ErrorCode != 0 {
		return fmt.Errorf("douyin TravelAgencyOrderConfirm 失败: error_code=%d description=%s order_id=%s",
			*result.Data.ErrorCode, tea.StringValue(result.Data.Description), tea.StringValue(result.Data.OrderID))
	}
	if result.Extra != nil && result.Extra.ErrorCode != nil && *result.Extra.ErrorCode != 0 {
		return fmt.Errorf("douyin TravelAgencyOrderConfirm 失败: error_code=%d description=%s logid=%s",
			*result.Extra.ErrorCode, tea.StringValue(result.Extra.Description), tea.StringValue(result.Extra.Logid))
	}
	return nil
}
