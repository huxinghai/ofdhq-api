package douyin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ofdhq-api/app/global/variable"
	"ofdhq-api/pkg/verify"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 抖音本地生活 SPI 回调签名校验（gin 中间件 + 请求验签入口）。
// 签名计算与校验算法在 SDK 风格的 pkg/verify 包中实现（API 与官方
// douyin-sign-verify-go/pkg/verify 示例一致），规则文档：
// https://partner.open-douyin.com/docs/resource/zh-CN/local-life/develop/preparation/spi-signature-rules

// VerifySPIRequest 校验抖音 SPI 回调请求的签名。
// 请求携带 header x-life-sign 时优先做新版 SHA-256 校验，
// 否则取 URL 参数 sign 做旧版 MD5 校验，两者都没有直接判失败。
// 校验过程会读取 body 并重新写回 r.Body，供后续业务绑定使用。
func VerifySPIRequest(r *http.Request, clientSecret string) error {
	if r.Header.Get("x-life-sign") != "" {
		if ok, expected, provided := verify.VerifyNewSignatureFromRequest(r, clientSecret); ok {
			return nil
		} else {
			return fmt.Errorf("douyin SPI 新版签名(x-life-sign)校验失败 expected=%s provided=%s", expected, provided)
		}
	}
	if r.URL.Query().Get("sign") != "" {
		if ok, expected, provided := verify.VerifyOldSignatureFromRequest(r, clientSecret); ok {
			return nil
		} else {
			return fmt.Errorf("douyin SPI 旧版签名(sign)校验失败 expected=%s provided=%s", expected, provided)
		}
	}
	return fmt.Errorf("douyin SPI 请求缺少签名(x-life-sign / sign 均为空)")
}

// dumpHeaders 汇总请求头为 "Key: Value; Key: Value" 形式，用于日志打印
func dumpHeaders(h http.Header) string {
	var sb strings.Builder
	for k, vs := range h {
		for _, v := range vs {
			if sb.Len() > 0 {
				sb.WriteString("; ")
			}
			sb.WriteString(k)
			sb.WriteString(": ")
			sb.WriteString(v)
		}
	}
	return sb.String()
}

// SpiSignVerify gin 中间件：记录请求日志 + 校验 SPI 回调签名 + client_key。
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

		// 打印回调的参数与头部信息，便于联调排查（读取后还原 body，不影响后续验签与绑定）
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			variable.ZapLog.Warn("douyin SPI 读取 body 失败", zap.Error(err))
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		variable.ZapLog.Info("douyin SPI 回调请求",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("client_ip", c.ClientIP()),
			zap.String("headers", dumpHeaders(c.Request.Header)),
			zap.String("body", string(body)),
		)

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
