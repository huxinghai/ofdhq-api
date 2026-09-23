package test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/bytedance/douyin-openapi-credential-go/client"
	"github.com/bytedance/douyin-openapi-sdk-go/client"
	util "github.com/bytedance/douyin-openapi-util-go/client"

	"ofdhq-api/app/global/variable"
	douyin "ofdhq-api/app/utils/douyin"
	"ofdhq-api/app/utils/yml_config"
)

// 解决默认-test.fullpath=true 导致的测试用例失败问题
func init() {
	flag.Bool("test.fullpath", false, "")
}

// app/utils/douyin 封装依赖全局配置指针（正常由 bootstrap 初始化），测试进程里手动初始化
func TestMain(m *testing.M) {
	variable.ConfigYml = yml_config.CreateYamlFactory()
	os.Exit(m.Run())
}

// 抖音开放平台 SDK 集成测试，调用真实接口，凭证通过环境变量提供，缺少时自动跳过：
//
//	DOUYIN_CLIENT_KEY / DOUYIN_CLIENT_SECRET   开放平台应用的 AppID(client key) / AppSecret
//	DOUYIN_OPEN_ID                             (视频相关接口需要) 授权用户的 open_id
//	DOUYIN_VIDEO_FILE                          (可选) 本地视频文件路径，用于上传+发布测试
//	DOUYIN_ORDER_ID                            (可选) 订单号，用于订单详情测试
//	DOUYIN_ACCOUNT_ID                          订单查询用 account_id（Hermes / Goodlife 交易订单查询必填）
//
// 运行示例：
//	DOUYIN_CLIENT_KEY=xx DOUYIN_CLIENT_SECRET=xx DOUYIN_OPEN_ID=xx DOUYIN_ACCESS_TOKEN=xx \
//		go test -v -run TestDouyin ./test/

func newDouyinClient(t *testing.T) *client.Client {
	key, secret := os.Getenv("DOUYIN_CLIENT_KEY"), os.Getenv("DOUYIN_CLIENT_SECRET")
	if key == "" || secret == "" {
		t.Skip("缺少环境变量 DOUYIN_CLIENT_KEY / DOUYIN_CLIENT_SECRET")
	}
	fmt.Printf("key:%v, secret:%v \n", key, secret)
	cli, err := client.NewClient(&credential.Config{
		ClientKey:    tea.String(key),
		ClientSecret: tea.String(secret),
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return cli
}

// 只用 AppID/AppSecret 自动获取 access_token（credential 内部缓存并按过期时间自动刷新）
// 注意用 GetClientToken（开放平台 open.douyin.com/oauth/client_token/）；
// GetAccessToken 打的是小程序端点 developer.toutiao.com，开放平台应用会报 40015 bad appid
func autoAccessToken(t *testing.T, cli *client.Client) string {
	token, err := cli.Credential.GetClientToken()
	if err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	accessToken := tea.StringValue(token.AccessToken)
	if accessToken == "" {
		t.Fatal("access_token 为空")
	}
	t.Logf("access_token 自动获取成功 accessToken:%v, expires_in=%d", accessToken, tea.Int64Value(token.ExpiresIn))
	return accessToken
}

// 视频类接口需要 open_id 标识授权用户，无法从 AppID/Secret 推导
func requireOpenID(t *testing.T) string {
	openID := os.Getenv("DOUYIN_OPEN_ID")
	if openID == "" {
		t.Skip("缺少环境变量 DOUYIN_OPEN_ID")
	}
	return openID
}

func checkResp(t *testing.T, action string, errorCode *int32, description *string) {
	t.Helper()
	if errorCode != nil && *errorCode != 0 {
		t.Fatalf("%s 失败: error_code=%d description=%s", action, *errorCode, tea.StringValue(description))
	}
	t.Logf("%s 成功: %s (logid 见 extra)", action, tea.StringValue(description))
}

// 1. 获取 client_token（只需 client key/secret，不需要用户授权）
func TestDouyinClientToken(t *testing.T) {
	cli := newDouyinClient(t)

	token, err := cli.Credential.GetClientToken()
	if err != nil {
		t.Fatalf("GetClientToken: %v", err)
	}
	if tea.StringValue(token.AccessToken) == "" {
		t.Fatal("client_token 为空")
	}
	t.Logf("client_token 获取成功, expires_in=%d", tea.Int64Value(token.ExpiresIn))
}

// 2. 查询授权账号的视频列表
func TestDouyinVideoList(t *testing.T) {
	cli := newDouyinClient(t)
	openID := requireOpenID(t)
	accessToken := autoAccessToken(t, cli)

	resp, err := cli.VideoVideoList(&client.VideoVideoListRequest{
		OpenId:      tea.String(openID),
		AccessToken: tea.String(accessToken),
		Count:       tea.Int32(10),
	})
	if err != nil {
		t.Fatalf("VideoVideoList: %v", err)
	}
	var code *int32
	var desc *string
	if resp.Extra != nil {
		code, desc = resp.Extra.ErrorCode, resp.Extra.Description
	}
	checkResp(t, "查询视频列表", code, desc)
	if resp.Data != nil {
		t.Logf("视频数量=%d has_more=%v", len(resp.Data.List), tea.BoolValue(resp.Data.HasMore))
		for i, v := range resp.Data.List {
			t.Logf("[%d] title=%s video_id=%s create_time=%d",
				i, tea.StringValue(v.Title), tea.StringValue(v.VideoId), tea.Int64Value(v.CreateTime))
		}
	}
}

// 4. 按时间范围查询订单（达人团购订单，最近 7 天）
func TestDouyinOrderQuery(t *testing.T) {
	cli := newDouyinClient(t)
	accessToken := autoAccessToken(t, cli)

	req := &client.OrderQueryRequest{
		AccessToken: tea.String(accessToken),
		StartTime:   tea.Int64(time.Now().Add(-7 * 24 * time.Hour).Unix()),
		EndTime:     tea.Int64(time.Now().Unix()),
		Page:        tea.Int32(1),
		Size:        tea.Int32(10),
		IsAsc:       tea.Bool(false),
	}
	// 可选：按达人 account_id 过滤
	if accountID := os.Getenv("DOUYIN_ACCOUNT_ID"); accountID != "" {
		req.AccountId = tea.String(accountID)
	}

	resp, err := cli.OrderQuery(req)
	if err != nil {
		t.Fatalf("OrderQuery: %v", err)
	}
	var code *int32
	var desc *string
	if resp.Extra != nil {
		code, desc = resp.Extra.ErrorCode, resp.Extra.Description
	}
	checkResp(t, "查询订单", code, desc)
	if resp.Data != nil {
		t.Logf("订单总数=%d", tea.Int32Value(resp.Data.Total))
		for i, o := range resp.Data.Orders {
			t.Logf("[%d] order_id=%s product=%s merchant=%s status=%d commission_ratio=%s create_time=%d",
				i, tea.StringValue(o.Id), tea.StringValue(o.ProductName), tea.StringValue(o.MerchantName),
				tea.IntValue(o.Status), tea.StringValue(o.CommissionRatio), tea.Int64Value(o.CreateTime))
		}
	}
}

// 5. 查询单个订单详情（需要 DOUYIN_ORDER_ID）
func TestDouyinOrderGet(t *testing.T) {
	cli := newDouyinClient(t)
	accessToken := autoAccessToken(t, cli)
	orderID := os.Getenv("DOUYIN_ORDER_ID")
	if orderID == "" {
		t.Skip("缺少环境变量 DOUYIN_ORDER_ID")
	}

	resp, err := cli.OrderGet(&client.OrderGetRequest{
		AccessToken: tea.String(accessToken),
		OrderId:     tea.String(orderID),
	})
	if err != nil {
		t.Fatalf("OrderGet: %v", err)
	}
	var code *int32
	var desc *string
	if resp.Extra != nil {
		code, desc = resp.Extra.ErrorCode, resp.Extra.Description
	}
	checkResp(t, "查询订单详情", code, desc)
	if resp.Data != nil {
		t.Logf("order_id=%s product=%s merchant=%s status=%d account_id=%s create_time=%d",
			tea.StringValue(resp.Data.Id), tea.StringValue(resp.Data.ProductName),
			tea.StringValue(resp.Data.MerchantName), tea.IntValue(resp.Data.Status),
			tea.StringValue(resp.Data.AccountId), tea.Int64Value(resp.Data.CreateTime))
	}
}

// 6. 查询 Hermes 交易订单（按 account_id + 下单时间范围分页查询，account_id 必填）
func TestDouyinHermesTradeOrderQuery(t *testing.T) {
	cli := newDouyinClient(t)
	accessToken := autoAccessToken(t, cli)
	accountID := os.Getenv("DOUYIN_ACCOUNT_ID")
	if accountID == "" {
		t.Skip("缺少环境变量 DOUYIN_ACCOUNT_ID")
	}

	startTime := time.Date(2026, 06, 01, 0, 0, 0, 0, time.Now().Location())

	resp, err := cli.HermesTradeOrderQuery(&client.HermesTradeOrderQueryRequest{
		AccessToken:          tea.String(accessToken),
		AccountId:            tea.String(accountID),
		CreateOrderStartTime: tea.Int64(startTime.Unix()),
		CreateOrderEndTime:   tea.Int64(startTime.AddDate(0, 1, 0).Unix()),
		PageNum:              tea.Int32(101),
		PageSize:             tea.Int32(10),
	})
	if err != nil {
		t.Fatalf("HermesTradeOrderQuery: %v", err)
	}
	var code *int32
	var desc *string
	if resp.Extra != nil {
		code, desc = resp.Extra.ErrorCode, resp.Extra.Description
	}
	checkResp(t, "查询 Hermes 交易订单", code, desc)
	if resp.Data != nil {
		if resp.Data.Page != nil {
			t.Logf("总订单数=%d page=%d/%d", tea.Int64Value(resp.Data.Page.Total),
				tea.Int32Value(resp.Data.Page.PageNum), tea.Int32Value(resp.Data.Page.PageSize))
		}
		for i, o := range resp.Data.Orders {
			t.Logf("[%d] order_id=%s pay_amount=%d sku数=%d create_time=%d pay_time=%d",
				i, tea.StringValue(o.OrderId), tea.Int32Value(o.PayAmount),
				tea.Int32Value(o.SkuNum), tea.Int64Value(o.CreateOrderTime), tea.Int64Value(o.PayTime))
		}
	}
}

// 7. 查询本地生活交易订单（goodlife /goodlife/v1/trade/order/query/，account_id 必填）
func TestDouyinTradeOrderQuery(t *testing.T) {
	cli := newDouyinClient(t)
	accessToken := autoAccessToken(t, cli)
	accountID := os.Getenv("DOUYIN_ACCOUNT_ID")
	if accountID == "" {
		t.Skip("缺少环境变量 DOUYIN_ACCOUNT_ID")
	}

	startTime := time.Date(2026, 06, 01, 0, 0, 0, 0, time.Now().Location())

	resp, err := cli.TradeOrderQuery(&client.TradeOrderQueryRequest{
		AccessToken:          tea.String(accessToken),
		AccountId:            tea.String(accountID),
		CreateOrderStartTime: tea.Int64(startTime.Unix()),
		CreateOrderEndTime:   tea.Int64(startTime.AddDate(0, 1, 0).Unix()),
		PageNum:              tea.Int32(1),
		PageSize:             tea.Int32(100),
		OrderStatus:          tea.Int32(1),
		UpdateOrderEndTime:   tea.Int64(1),
		UpdateOrderStartTime: tea.Int64(1),
	})
	if err != nil {
		t.Fatalf("TradeOrderQuery: %v", err)
	}
	var code *int32
	var desc *string
	if resp.Extra != nil {
		code, desc = resp.Extra.ErrorCode, resp.Extra.Description
	}
	checkResp(t, "查询交易订单", code, desc)
	if resp.Data == nil || len(resp.Data.Orders) == 0 {
		// 空数据时打印完整响应便于排查（配合 DEBUG=tea 可看请求 URL / 响应状态）
		t.Logf("未返回订单, 完整响应: %s", resp.GoString())
		return
	}
	if resp.Data.Page != nil {
		t.Logf("总订单数=%d page=%d/%d", tea.Int64Value(resp.Data.Page.Total),
			tea.Int32Value(resp.Data.Page.PageNum), tea.Int32Value(resp.Data.Page.PageSize))
	}
	for i, o := range resp.Data.Orders {
		t.Logf("[%d] order_id=%s sku=%s pay_amount=%d status=%d count=%d create_time=%d pay_time=%d",
			i, tea.StringValue(o.OrderId), tea.StringValue(o.SkuName), tea.Int32Value(o.PayAmount),
			tea.Int32Value(o.OrderStatus), tea.Int32Value(o.Count),
			tea.Int64Value(o.CreateOrderTime), tea.Int64Value(o.PayTime))
	}
}

// 8. 直连调用交易订单查询（绕过 SDK）：
// SDK 的 TradeOrderQuery 会把未设置的字段以零值拼进 query —— nil *int64 经 tea.Int64Value
// 变成 0，StringifyMapValue 只过滤 nil，因此 URL 固定带 update_order_start_time=0&
// update_order_end_time=0，与 create_order_* 时间窗冲突导致查不到数据。
// 这里自己拼 query，只发送显式设置的参数。
func TestDouyinTradeOrderQueryDirect(t *testing.T) {
	cli := newDouyinClient(t)
	accessToken := autoAccessToken(t, cli)
	accountID := os.Getenv("DOUYIN_ACCOUNT_ID")
	if accountID == "" {
		t.Skip("缺少环境变量 DOUYIN_ACCOUNT_ID")
	}

	startTime := time.Date(2026, 06, 01, 0, 0, 0, 0, time.Now().Location())

	q := url.Values{}
	q.Set("account_id", accountID)
	q.Set("create_order_start_time", strconv.FormatInt(startTime.Unix(), 10))
	q.Set("create_order_end_time", strconv.FormatInt(startTime.AddDate(0, 1, 0).Unix(), 10))
	q.Set("page_num", "1")
	q.Set("page_size", "10")

	reqURL := "https://open.douyin.com/goodlife/v1/trade/order/query/?" + q.Encode()
	t.Logf("请求 URL: %s", reqURL)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("access-token", accessToken)

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	t.Logf("HTTP %d, body=%s", httpResp.StatusCode, body)

	var resp client.TradeOrderQueryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	var code *int32
	var desc *string
	if resp.Extra != nil {
		code, desc = resp.Extra.ErrorCode, resp.Extra.Description
	}
	checkResp(t, "直连查询交易订单", code, desc)
	for i, o := range resp.Data.Orders {
		t.Logf("[%d] order_id=%s sku=%s pay_amount=%d status=%d count=%d create_time=%d pay_time=%d",
			i, tea.StringValue(o.OrderId), tea.StringValue(o.SkuName), tea.Int32Value(o.PayAmount),
			tea.Int32Value(o.OrderStatus), tea.Int32Value(o.Count),
			tea.Int64Value(o.CreateOrderTime), tea.Int64Value(o.PayTime))
	}
}

// 9. 走 app/utils/douyin 封装查询交易订单：
// 配置读 config/config.yml 的 Douyin.ClientKey/ClientSecret/AccountId，
// access_token 走 Redis 缓存（未命中自动获取并回写）。
// 将真实凭证填入 config.yml 后运行：go test -v -run TestDouyinTradeOrderQueryUtil ./test/
func TestDouyinTradeOrderQueryUtil(t *testing.T) {
	if variable.ConfigYml.GetString("Douyin.ClientKey") == "" ||
		variable.ConfigYml.GetString("Douyin.ClientKey") == "douyin_client_key" {
		t.Skip("config.yml 未配置真实的 Douyin.ClientKey")
	}

	startTime := time.Date(2026, 06, 01, 0, 0, 0, 0, time.Now().Location())
	resp, err := douyin.TradeOrderQuery(&douyin.TradeOrderQueryParam{
		CreateOrderStartTime: startTime.Unix(),
		CreateOrderEndTime:   startTime.AddDate(0, 1, 0).Unix(),
		PageNum:              1,
		PageSize:             10,
	})
	if err != nil {
		t.Fatalf("douyin.TradeOrderQuery: %v", err)
	}
	var code *int32
	var desc *string
	if resp.Extra != nil {
		code, desc = resp.Extra.ErrorCode, resp.Extra.Description
	}
	checkResp(t, "封装查询交易订单", code, desc)
	if resp.Data != nil && resp.Data.Page != nil {
		t.Logf("总订单数=%d page=%d/%d", tea.Int64Value(resp.Data.Page.Total),
			tea.Int32Value(resp.Data.Page.PageNum), tea.Int32Value(resp.Data.Page.PageSize))
	}
	for i, o := range resp.Data.Orders {
		t.Logf("[%d] order_id=%s sku=%s pay_amount=%d status=%d count=%d create_time=%d pay_time=%d",
			i, tea.StringValue(o.OrderId), tea.StringValue(o.SkuName), tea.Int32Value(o.PayAmount),
			tea.Int32Value(o.OrderStatus), tea.Int32Value(o.Count),
			tea.Int64Value(o.CreateOrderTime), tea.Int64Value(o.PayTime))
	}
}

// 3. 上传视频并发布（需要 DOUYIN_VIDEO_FILE 指向本地视频文件）
func TestDouyinUploadAndCreateVideo(t *testing.T) {
	cli := newDouyinClient(t)
	openID := requireOpenID(t)
	accessToken := autoAccessToken(t, cli)
	videoFile := os.Getenv("DOUYIN_VIDEO_FILE")
	if videoFile == "" {
		t.Skip("缺少环境变量 DOUYIN_VIDEO_FILE")
	}

	f, err := os.Open(videoFile)
	if err != nil {
		t.Fatalf("打开视频文件: %v", err)
	}
	defer f.Close()

	// 3.1 上传视频源文件
	up, err := cli.VideoUploadVideo(&client.VideoUploadVideoRequest{
		OpenId:      tea.String(openID),
		AccessToken: tea.String(accessToken),
		Video: &util.FileField{
			Filename:    tea.String(filepath.Base(videoFile)),
			ContentType: tea.String("video/mp4"),
			Content:     f,
		},
	})
	if err != nil {
		t.Fatalf("VideoUploadVideo: %v", err)
	}
	var code *int32
	var desc *string
	if up.Extra != nil {
		code, desc = up.Extra.ErrorCode, up.Extra.Description
	}
	checkResp(t, "上传视频", code, desc)
	videoID := tea.StringValue(up.Data.Video.VideoId)
	if videoID == "" {
		t.Fatalf("上传成功但未返回 video_id: %s", up.Data.GoString())
	}
	t.Logf("video_id=%s 宽x高=%dx%d", videoID, tea.Int32Value(up.Data.Video.Width), tea.Int32Value(up.Data.Video.Height))

	// 3.2 用 video_id 创建（发布）视频
	cr, err := cli.VideoCreateVideo(&client.VideoCreateVideoRequest{
		OpenId:      tea.String(openID),
		AccessToken: tea.String(accessToken),
		VideoId:     tea.String(videoID),
		Text:        tea.String("SDK 测试发布"),
	})
	if err != nil {
		t.Fatalf("VideoCreateVideo: %v", err)
	}
	if cr.Extra != nil {
		code, desc = cr.Extra.ErrorCode, cr.Extra.Description
	}
	checkResp(t, "发布视频", code, desc)
	if cr.Data != nil {
		t.Logf("item_id=%s", tea.StringValue(cr.Data.ItemId))
	}
}

func TestAesDecrypt(t *testing.T) {
	// {
	// 	"name": "",
	// 	"phone": "149****7100",
	// 	"phone_encrypt": "MDUGJWk2ctOamW3gwEd4YA=="
	// }
	phone, err := AesDecrypt("MDUGJWk2ctOamW3gwEd4YA==", "30d07a13646ecb6d5b5604fa6d1203ee")
	fmt.Printf("电话号码:%v, %v \n", string(phone), err)
}

// AesDecrypt 解密函数
// encryptedStr：base64后的密文
// secret：appid/client_key对应的client_secret
// return: []byte 明文
func AesDecrypt(encryptedStr string, secret string) ([]byte, error) {
	// 加密字符串进行base64解码
	decodeBytes, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return nil, err
	}
	key, iv := parseSecret(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(decodeBytes))
	blockMode.CryptBlocks(origData, decodeBytes)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

// parseSecret 将secret解析为key和iv
func parseSecret(secret string) ([]byte, []byte) {
	// secret对齐为32位
	secret = cutSecret(secret)
	secret = fillSecret(secret)
	key, iv := secret, secret[16:]
	return []byte(key), []byte(iv)
}
func fillSecret(secret string) string {
	if len(secret) >= 32 {
		return secret
	}
	rightCnt := (32 - len(secret)) / 2
	leftCnt := 32 - len(secret) - rightCnt
	var byt bytes.Buffer
	byt.Write(bytes.Repeat([]byte("#"), leftCnt))
	byt.WriteString(secret)
	byt.Write(bytes.Repeat([]byte("#"), rightCnt))
	return byt.String()
}
func cutSecret(secret string) string {
	if len(secret) <= 32 {
		return secret
	}
	rightCnt := (len(secret) - 32) / 2
	leftCnt := len(secret) - 32 - rightCnt
	return secret[leftCnt : 32+leftCnt]
}
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
