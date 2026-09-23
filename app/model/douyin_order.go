package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"ofdhq-api/app/global/variable"

	"go.uber.org/zap"
)

// 抖音本地生活 SPI 回调落库模型。
// 设计约定：核心字段（订单号、金额、状态等业务检索字段）建独立列，
// 非核心字段整体以原始报文存 extra JSON 列，不逐个展开。

// 订单状态（由取消/退款通知更新）
const (
	DouyinOrderStatusCreated  int16 = 1 // 已创建
	DouyinOrderStatusCanceled int16 = 2 // 已取消
	DouyinOrderStatusRefunded int16 = 3 // 已退款
)

// 通知类型
const (
	DouyinNoticeTypeCancel int16 = 1 // 预售券订单取消通知(travel_spot.order.cancel_apply)
	DouyinNoticeTypeRefund int16 = 2 // 预售券退款通知(travel_spot.order.refund_notify)
)

func douyinNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// genOrderOutId 生成我方订单号（创单成功必返给抖音），使用雪花ID
func genOrderOutId() string {
	if variable.SnowFlake == nil {
		return ""
	}
	return strconv.FormatInt(variable.SnowFlake.GetId(), 10)
}

// ------------------------- 预售订单（biz_type 3011） -------------------------

type DouyinPresaleOrderModel struct {
	BaseModel
	OrderId                string          `json:"order_id"`                // 抖音侧预售订单号
	OrderOutId             string          `json:"order_out_id"`            // 我方订单号（创单成功必返给抖音）
	BizType                int32           `json:"biz_type"`                // 3011 旅行社预售券
	CreateOrderTimeUnix    int64           `json:"create_order_time_unix"`  // 下单时间（秒）
	TotalAmount            int64           `json:"total_amount"`            // 总原始金额（分）
	TotalCouponCount       int32           `json:"total_coupon_count"`      // 券总张数（预留，一单一券）
	EachCouponAmount       int64           `json:"each_coupon_amount"`      // 单张券原始金额（分）
	PayAmount              int64           `json:"pay_amount"`              // 用户实付（分）
	ActualAmount           int64           `json:"actual_amount"`           // 实收=实付+平台补贴（分）
	DiscountAmount         int64           `json:"discount_amount"`         // 优惠总金额（分）
	MerchantDiscountAmount int64           `json:"merchant_discount_amount"` // 商家优惠金额（分）
	BuyerPhone             string          `json:"buyer_phone"`             // 买家手机号（加密）
	PayTimeUnix            int64           `json:"pay_time_unix"`           // 支付时间（秒），支付后创单模式必传
	OrderSource            int16           `json:"order_source"`            // 1抖音 2抖省省 3豆包
	Status                 int16           `json:"status"`                  // 1已创建 2已取消 3已退款
	Extra                  json.RawMessage `json:"extra"`                   // 非核心字段原始报文
}

func CreateDouyinPresaleOrderFactory() *DouyinPresaleOrderModel {
	return &DouyinPresaleOrderModel{BaseModel: BaseModel{DB: UseDbConn("")}}
}

func (t *DouyinPresaleOrderModel) TableName() string {
	return "douyin_presale_orders"
}

// FindByOrderId 按抖音订单号查询（幂等重试时复用已有记录）
func (t *DouyinPresaleOrderModel) FindByOrderId(orderId string) (*DouyinPresaleOrderModel, error) {
	var row DouyinPresaleOrderModel
	result := t.Raw("SELECT * FROM `douyin_presale_orders` WHERE `order_id` = ? LIMIT 1", orderId).First(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	return &row, nil
}

// CreateOrGet 幂等创单：order_id 已存在时返回已有记录（created=false），否则插入新记录
func (t *DouyinPresaleOrderModel) CreateOrGet(order *DouyinPresaleOrderModel) (*DouyinPresaleOrderModel, bool, error) {
	if order.Status == 0 {
		order.Status = DouyinOrderStatusCreated
	}
	if order.OrderOutId == "" {
		order.OrderOutId = genOrderOutId()
	}
	now := douyinNow()
	order.CreatedAt = now
	order.UpdatedAt = now
	err := t.Create(order).Error
	if err == nil {
		return order, true, nil
	}
	// 插入失败大概率是 order_id 唯一键冲突（抖音超时重试），回查已有记录
	exist, qErr := t.FindByOrderId(order.OrderId)
	if qErr != nil {
		return nil, false, errors.Join(err, fmt.Errorf("DouyinPresaleOrderModel.CreateOrGet 冲突回查失败: %s", order.OrderId))
	}
	return exist, false, nil
}

// UpdateStatusByOrderId 取消/退款通知更新订单状态
func (t *DouyinPresaleOrderModel) UpdateStatusByOrderId(orderId string, status int16) error {
	return t.Exec("UPDATE `douyin_presale_orders` SET `status` = ? WHERE `order_id` = ?", status, orderId).Error
}

// ------------------------- 预约订单（biz_type 3012） -------------------------

type DouyinBookOrderModel struct {
	BaseModel
	OrderId                string          `json:"order_id"`                 // 抖音侧预约订单号
	OrderOutId             string          `json:"order_out_id"`             // 我方订单号（创单成功必返给抖音）
	BizType                int32           `json:"biz_type"`                 // 3012 旅行社预约单
	SourceOrderId          string          `json:"source_order_id"`          // 归属的预售订单 ID
	PresaleCouponId        string          `json:"presale_coupon_id"`        // 预售券 ID
	CreateOrderTimeUnix    int64           `json:"create_order_time_unix"`   // 创单时间（秒）
	PayAmount              int64           `json:"pay_amount"`               // 用户实付（分）
	ActualAmount           int64           `json:"actual_amount"`            // 实收（分）
	OriginalAmount         int64           `json:"original_amount"`          // 原始金额（分）
	DiscountAmount         int64           `json:"discount_amount"`          // 优惠总金额（分）
	MerchantDiscountAmount int64           `json:"merchant_discount_amount"` // 商家优惠金额（分）
	PayTimeUnix            int64           `json:"pay_time_unix"`            // 支付时间（秒），支付后创单模式必传
	OrderSource            int16           `json:"order_source"`             // 1抖音 2抖省省 3豆包
	BookStartDate          string          `json:"book_start_date"`          // 预约开始日期 yyyy-MM-dd
	BookEndDate            string          `json:"book_end_date"`            // 预约结束日期 yyyy-MM-dd
	OccupantCount          int32           `json:"occupant_count"`           // 入住人数（book_info.occupancies 数量）
	Status                 int16           `json:"status"`                   // 1已创建 2已取消 3已退款
	Extra                  json.RawMessage `json:"extra"`                    // 非核心字段原始报文
}

func CreateDouyinBookOrderFactory() *DouyinBookOrderModel {
	return &DouyinBookOrderModel{BaseModel: BaseModel{DB: UseDbConn("")}}
}

func (t *DouyinBookOrderModel) TableName() string {
	return "douyin_book_orders"
}

func (t *DouyinBookOrderModel) FindByOrderId(orderId string) (*DouyinBookOrderModel, error) {
	var row DouyinBookOrderModel
	result := t.Raw("SELECT * FROM `douyin_book_orders` WHERE `order_id` = ? LIMIT 1", orderId).First(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	return &row, nil
}

// CreateOrGet 幂等创单：order_id 已存在时返回已有记录（created=false），否则插入新记录
func (t *DouyinBookOrderModel) CreateOrGet(order *DouyinBookOrderModel) (*DouyinBookOrderModel, bool, error) {
	if order.Status == 0 {
		order.Status = DouyinOrderStatusCreated
	}
	if order.OrderOutId == "" {
		order.OrderOutId = genOrderOutId()
	}
	now := douyinNow()
	order.CreatedAt = now
	order.UpdatedAt = now
	err := t.Create(order).Error
	if err == nil {
		return order, true, nil
	}
	exist, qErr := t.FindByOrderId(order.OrderId)
	if qErr != nil {
		return nil, false, errors.Join(err, fmt.Errorf("DouyinBookOrderModel.CreateOrGet 冲突回查失败: %s", order.OrderId))
	}
	return exist, false, nil
}

func (t *DouyinBookOrderModel) UpdateStatusByOrderId(orderId string, status int16) error {
	return t.Exec("UPDATE `douyin_book_orders` SET `status` = ? WHERE `order_id` = ?", status, orderId).Error
}

// --------------------- 取消/退款通知流水（SPI 通知类） ---------------------

type DouyinOrderNoticeModel struct {
	BaseModel
	DedupKey       string          `json:"dedup_key"`       // 幂等键
	NoticeType     int16           `json:"notice_type"`     // 1取消通知 2退款通知
	OrderId        string          `json:"order_id"`        // 抖音侧订单号
	OrderOutId     string          `json:"order_out_id"`    // 第三方订单 ID
	BizType        int32           `json:"biz_type"`        // 取消通知携带：3011预售券 3012预约单
	SubType        int16           `json:"sub_type"`        // 取消:1支付前2支付后3外部原因；退款:1订单退款2补差价退款
	NotifyTimeUnix int64           `json:"notify_time_unix"` // 取消:cancel_order_time_unix；退款:refund_time_unix（秒）
	PayAmount      int64           `json:"pay_amount"`      // 订单实付金额（分，退款通知）
	RefundAmount   int64           `json:"refund_amount"`   // 实际退款金额（分，退款通知）
	AfterSaleId    string          `json:"after_sale_id"`   // 售后单 ID（退款通知）
	Extra          json.RawMessage `json:"extra"`           // 非核心字段原始报文
}

func CreateDouyinOrderNoticeFactory() *DouyinOrderNoticeModel {
	return &DouyinOrderNoticeModel{BaseModel: BaseModel{DB: UseDbConn("")}}
}

func (t *DouyinOrderNoticeModel) TableName() string {
	return "douyin_order_notices"
}

// SaveDedup 按幂等键落库通知流水：dedup_key 已存在时不重复插入（返回 false）
func (t *DouyinOrderNoticeModel) SaveDedup(notice *DouyinOrderNoticeModel) (bool, error) {
	now := douyinNow()
	notice.CreatedAt = now
	notice.UpdatedAt = now
	result := t.Exec("INSERT IGNORE INTO `douyin_order_notices` (`dedup_key`,`notice_type`,`order_id`,`order_out_id`,`biz_type`,`sub_type`,`notify_time_unix`,`pay_amount`,`refund_amount`,`after_sale_id`,`extra`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)",
		notice.DedupKey, notice.NoticeType, notice.OrderId, notice.OrderOutId, notice.BizType, notice.SubType,
		notice.NotifyTimeUnix, notice.PayAmount, notice.RefundAmount, notice.AfterSaleId, string(notice.Extra), now, now)
	if result.Error != nil {
		return false, errors.Join(result.Error, fmt.Errorf("DouyinOrderNoticeModel.SaveDedup 落库失败: %s", notice.DedupKey))
	}
	return result.RowsAffected > 0, nil
}

// UpdateOrderStatusByOrderId 取消/退款通知联动更新订单状态（预约单/预售单哪个存在更新哪个）
func UpdateDouyinOrderStatusByOrderId(orderId string, status int16) {
	if book := CreateDouyinBookOrderFactory(); book.DB != nil {
		if err := book.UpdateStatusByOrderId(orderId, status); err != nil {
			variable.ZapLog.Warn("douyin 通知联动更新预约订单状态失败", zap.String("order_id", orderId), zap.Error(err))
		}
	}
	if presale := CreateDouyinPresaleOrderFactory(); presale.DB != nil {
		if err := presale.UpdateStatusByOrderId(orderId, status); err != nil {
			variable.ZapLog.Warn("douyin 通知联动更新预售订单状态失败", zap.String("order_id", orderId), zap.Error(err))
		}
	}
}
