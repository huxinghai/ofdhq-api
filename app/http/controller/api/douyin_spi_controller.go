package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DouyinSpi 抖音本地生活 SPI 回调（抖音 -> 本系统），出参统一为 {"data":{error_code,...}}。
// 按约定只实现/校验各 SPI 的必填字段，非必填字段不逐个绑定；
// 完整原始报文整体存入各表 extra JSON 列。
// error_code 语义（创单类）：0成功；1-8拒单不重试；100重试；其他默认重试。
type DouyinSpi struct{}

func spiResp(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// readSpiBody 读取回调原始 body（验签中间件已还原），供绑定与 extra 落库复用
func readSpiBody(c *gin.Context) ([]byte, error) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return []byte("{}"), nil
	}
	return raw, nil
}

// ------------------- SPI #5 预售券创建预售订单 travel_spot.order.create_presale_order -------------------

type douyinPresaleBuyerInfo struct {
	Phone string `json:"phone"` // 加密
}

type douyinPresalePayInfo struct {
	PayTimeUnix int64 `json:"pay_time_unix"` // 秒
	OrderSource int16 `json:"order_source"`  // 1抖音 2抖省省 3豆包
}

type douyinCreatePresaleOrderReq struct {
	// 必填字段
	OrderId                string                   `json:"order_id"`
	BizType                int32                    `json:"biz_type"`
	CreateOrderTimeUnix    int64                    `json:"create_order_time_unix"`
	TotalAmount            int64                    `json:"total_amount"`
	TotalCouponCount       int32                    `json:"total_coupon_count"`
	EachCouponAmount       int64                    `json:"each_coupon_amount"`
	PayAmount              int64                    `json:"pay_amount"`
	ActualAmount           int64                    `json:"actual_amount"`
	DiscountAmount         int64                    `json:"discount_amount"`
	MerchantDiscountAmount int64                    `json:"merchant_discount_amount"`
	BuyerInfo              douyinPresaleBuyerInfo   `json:"buyer_info"`
	DepartureInfo          string                   `json:"departure_info"`
	// 支付后创单模式必传（支付前创单不传），此处不强制
	PayInfo douyinPresalePayInfo `json:"pay_info"`
}

func (t *DouyinSpi) CreatePresaleOrder(c *gin.Context) {
	raw, err := readSpiBody(c)
	if err != nil {
		spiResp(c, gin.H{"error_code": 4000002, "description": "read body failed"})
		return
	}
	var req douyinCreatePresaleOrderReq
	if err := json.Unmarshal(raw, &req); err != nil {
		variable.ZapLog.Warn("douyin SPI 创建预售订单参数解析失败", zap.Error(err))
		spiResp(c, gin.H{"error_code": 4000002, "description": "invalid json body"})
		return
	}
	if req.OrderId == "" || req.BizType == 0 || req.CreateOrderTimeUnix == 0 {
		spiResp(c, gin.H{"error_code": 4000002, "description": "order_id/biz_type/create_order_time_unix required"})
		return
	}

	order := &model.DouyinPresaleOrderModel{
		OrderId:                req.OrderId,
		BizType:                req.BizType,
		CreateOrderTimeUnix:    req.CreateOrderTimeUnix,
		TotalAmount:            req.TotalAmount,
		TotalCouponCount:       req.TotalCouponCount,
		EachCouponAmount:       req.EachCouponAmount,
		PayAmount:              req.PayAmount,
		ActualAmount:           req.ActualAmount,
		DiscountAmount:         req.DiscountAmount,
		MerchantDiscountAmount: req.MerchantDiscountAmount,
		BuyerPhone:             req.BuyerInfo.Phone,
		PayTimeUnix:            req.PayInfo.PayTimeUnix,
		OrderSource:            req.PayInfo.OrderSource,
		Status:                 model.DouyinOrderStatusCreated,
		Extra:                  json.RawMessage(raw),
	}

	saved, created, err := model.CreateDouyinPresaleOrderFactory().CreateOrGet(order)
	if err != nil {
		variable.ZapLog.Error("douyin SPI 创建预售订单落库失败", zap.String("order_id", req.OrderId), zap.Error(err))
		// 系统错误返回 100，抖音侧会重试
		spiResp(c, gin.H{"error_code": 100, "description": "internal error", "order_id": req.OrderId})
		return
	}
	if !created {
		variable.ZapLog.Info("douyin SPI 创建预售订单幂等命中，复用已有订单", zap.String("order_id", req.OrderId))
	}
	spiResp(c, gin.H{
		"error_code":   0,
		"description":  "success",
		"order_id":     req.OrderId,
		"order_out_id": saved.OrderOutId,
	})
}

// ------------------- SPI #1 预售券创建预约订单 travel_spot.order.create_order -------------------

type douyinOccupancy struct {
	Name string `json:"name"`
}

type douyinBookInfo struct {
	BookStartDate string              `json:"book_start_date"` // yyyy-MM-dd
	BookEndDate   string              `json:"book_end_date"`   // yyyy-MM-dd
	Occupancies   []douyinOccupancy   `json:"occupancies"`     // 必填，入住人列表
}

type douyinBookPayInfo struct {
	PayTimeUnix int64 `json:"pay_time_unix"`
	OrderSource int16 `json:"order_source"`
}

type douyinCreateBookOrderReq struct {
	// 必填字段
	OrderId                string             `json:"order_id"`
	BizType                int32              `json:"biz_type"`
	CreateOrderTimeUnix    int64              `json:"create_order_time_unix"`
	PayAmount              int64              `json:"pay_amount"`
	ActualAmount           int64              `json:"actual_amount"`
	OriginalAmount         int64              `json:"original_amount"`
	DiscountAmount         int64              `json:"discount_amount"`
	MerchantDiscountAmount int64              `json:"merchant_discount_amount"`
	BookInfo               douyinBookInfo     `json:"book_info"`
	// 预售券预约关联（预约单核心检索字段，非必填不强制）
	SourceOrderId   string            `json:"source_order_id"`
	PresaleCouponId string            `json:"presale_coupon_id"`
	PayInfo         douyinBookPayInfo `json:"pay_info"`
}

func (t *DouyinSpi) CreateBookOrder(c *gin.Context) {
	raw, err := readSpiBody(c)
	if err != nil {
		spiResp(c, gin.H{"error_code": 4000002, "description": "read body failed"})
		return
	}
	var req douyinCreateBookOrderReq
	if err := json.Unmarshal(raw, &req); err != nil {
		variable.ZapLog.Warn("douyin SPI 创建预约订单参数解析失败", zap.Error(err))
		spiResp(c, gin.H{"error_code": 4000002, "description": "invalid json body"})
		return
	}
	if req.OrderId == "" || req.BizType == 0 || req.CreateOrderTimeUnix == 0 || len(req.BookInfo.Occupancies) == 0 {
		spiResp(c, gin.H{"error_code": 4000002, "description": "order_id/biz_type/create_order_time_unix/book_info.occupancies required"})
		return
	}

	order := &model.DouyinBookOrderModel{
		OrderId:                req.OrderId,
		BizType:                req.BizType,
		SourceOrderId:          req.SourceOrderId,
		PresaleCouponId:        req.PresaleCouponId,
		CreateOrderTimeUnix:    req.CreateOrderTimeUnix,
		PayAmount:              req.PayAmount,
		ActualAmount:           req.ActualAmount,
		OriginalAmount:         req.OriginalAmount,
		DiscountAmount:         req.DiscountAmount,
		MerchantDiscountAmount: req.MerchantDiscountAmount,
		PayTimeUnix:            req.PayInfo.PayTimeUnix,
		OrderSource:            req.PayInfo.OrderSource,
		BookStartDate:          req.BookInfo.BookStartDate,
		BookEndDate:            req.BookInfo.BookEndDate,
		OccupantCount:          int32(len(req.BookInfo.Occupancies)),
		Status:                 model.DouyinOrderStatusCreated,
		Extra:                  json.RawMessage(raw),
	}

	saved, created, err := model.CreateDouyinBookOrderFactory().CreateOrGet(order)
	if err != nil {
		variable.ZapLog.Error("douyin SPI 创建预约订单落库失败", zap.String("order_id", req.OrderId), zap.Error(err))
		spiResp(c, gin.H{"error_code": 100, "description": "internal error", "order_id": req.OrderId})
		return
	}
	if !created {
		variable.ZapLog.Info("douyin SPI 创建预约订单幂等命中，复用已有订单", zap.String("order_id", req.OrderId))
	}
	spiResp(c, gin.H{
		"error_code":   0,
		"description":  "success",
		"order_id":     req.OrderId,
		"order_out_id": saved.OrderOutId,
	})
}

// ------------------- SPI #3 预售券订单取消通知 travel_spot.order.cancel_apply -------------------

type douyinOrderCancelReq struct {
	// 必填字段
	OrderId             string `json:"order_id"`
	OrderOutId          string `json:"order_out_id"`
	CancelOrderTimeUnix int64  `json:"cancel_order_time_unix"` // 秒
	CancelType          int16  `json:"cancel_type"`            // 1支付前取消 2支付后取消 3外部原因(如创单失败)
	BizType             int32  `json:"biz_type"`               // 3011预售券 3012预约单
}

func (t *DouyinSpi) OrderCancel(c *gin.Context) {
	raw, err := readSpiBody(c)
	if err != nil {
		spiResp(c, gin.H{"error_code": 4000002, "description": "read body failed"})
		return
	}
	var req douyinOrderCancelReq
	if err := json.Unmarshal(raw, &req); err != nil {
		variable.ZapLog.Warn("douyin SPI 取消通知参数解析失败", zap.Error(err))
		spiResp(c, gin.H{"error_code": 4000002, "description": "invalid json body"})
		return
	}
	if req.OrderId == "" || req.OrderOutId == "" || req.CancelOrderTimeUnix == 0 || req.CancelType == 0 || req.BizType == 0 {
		spiResp(c, gin.H{"error_code": 4000002, "description": "order_id/order_out_id/cancel_order_time_unix/cancel_type/biz_type required"})
		return
	}

	notice := &model.DouyinOrderNoticeModel{
		DedupKey:       "cancel:" + req.OrderId + ":" + strconv.FormatInt(req.CancelOrderTimeUnix, 10),
		NoticeType:     model.DouyinNoticeTypeCancel,
		OrderId:        req.OrderId,
		OrderOutId:     req.OrderOutId,
		BizType:        req.BizType,
		SubType:        req.CancelType,
		NotifyTimeUnix: req.CancelOrderTimeUnix,
		Extra:          json.RawMessage(raw),
	}
	if _, err := model.CreateDouyinOrderNoticeFactory().SaveDedup(notice); err != nil {
		variable.ZapLog.Error("douyin SPI 取消通知落库失败", zap.String("order_id", req.OrderId), zap.Error(err))
		spiResp(c, gin.H{"error_code": 100, "description": "internal error"})
		return
	}
	// 联动更新订单状态
	model.UpdateDouyinOrderStatusByOrderId(req.OrderId, model.DouyinOrderStatusCanceled)
	spiResp(c, gin.H{"error_code": 0, "description": "success"})
}

// ------------------- SPI #4 预售券退款通知 travel_spot.order.refund_notify -------------------

type douyinRefundNotifyReq struct {
	// 必填字段
	OrderId        string `json:"order_id"`
	OrderOutId     string `json:"order_out_id"`
	PayAmount      int64  `json:"pay_amount"`       // 订单实付（分）
	RefundAmount   int64  `json:"refund_amount"`    // 实际退款金额（分）
	RefundTimeUnix int64  `json:"refund_time_unix"` // 秒
	RefundType     int16  `json:"refund_type"`      // 1订单退款 2补差价退款
	// 非必填
	AfterSaleId string `json:"after_sale_id"` // 售后单ID（幂等键组成部分）
}

func (t *DouyinSpi) OrderRefundNotify(c *gin.Context) {
	raw, err := readSpiBody(c)
	if err != nil {
		spiResp(c, gin.H{"error_code": 4000002, "description": "read body failed"})
		return
	}
	var req douyinRefundNotifyReq
	if err := json.Unmarshal(raw, &req); err != nil {
		variable.ZapLog.Warn("douyin SPI 退款通知参数解析失败", zap.Error(err))
		spiResp(c, gin.H{"error_code": 4000002, "description": "invalid json body"})
		return
	}
	if req.OrderId == "" || req.OrderOutId == "" || req.RefundTimeUnix == 0 || req.RefundType == 0 {
		spiResp(c, gin.H{"error_code": 4000002, "description": "order_id/order_out_id/refund_time_unix/refund_type required"})
		return
	}

	notice := &model.DouyinOrderNoticeModel{
		DedupKey:       "refund:" + req.OrderId + ":" + req.AfterSaleId + ":" + strconv.FormatInt(req.RefundTimeUnix, 10),
		NoticeType:     model.DouyinNoticeTypeRefund,
		OrderId:        req.OrderId,
		OrderOutId:     req.OrderOutId,
		SubType:        req.RefundType,
		NotifyTimeUnix: req.RefundTimeUnix,
		PayAmount:      req.PayAmount,
		RefundAmount:   req.RefundAmount,
		AfterSaleId:    req.AfterSaleId,
		Extra:          json.RawMessage(raw),
	}
	// 文档明确：退款通知主要用作通知，返回结果不影响最终退款结果，
	// 非超时情况下不要返回业务错误和系统错误 —— 落库失败仅记日志，仍返回成功
	if _, err := model.CreateDouyinOrderNoticeFactory().SaveDedup(notice); err != nil {
		variable.ZapLog.Error("douyin SPI 退款通知落库失败", zap.String("order_id", req.OrderId), zap.Error(err))
	}
	model.UpdateDouyinOrderStatusByOrderId(req.OrderId, model.DouyinOrderStatusRefunded)
	spiResp(c, gin.H{"error_code": 0, "description": "success"})
}
