package douyin

import (
	"errors"
	"fmt"
	"sync"

	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/utils/redis_factory"

	credential "github.com/bytedance/douyin-openapi-credential-go/client"
	"github.com/bytedance/douyin-openapi-sdk-go/client"
	"github.com/alibabacloud-go/tea/tea"
	"go.uber.org/zap"
)

// SDK 客户端懒加载单例，只用于获取 access_token（credential 内部有进程内缓存并按过期时间自动刷新）
var (
	sdkClient *client.Client
	sdkOnce   sync.Once
	sdkErr    error
)

func newSDKClient() (*client.Client, error) {
	sdkOnce.Do(func() {
		sdkClient, sdkErr = client.NewClient(&credential.Config{
			ClientKey:    tea.String(variable.ConfigYml.GetString("Douyin.ClientKey")),
			ClientSecret: tea.String(variable.ConfigYml.GetString("Douyin.ClientSecret")),
		})
	})
	return sdkClient, sdkErr
}

// GetAccessToken 获取开放平台 access_token（client_token），优先读 Redis 缓存，未命中时走 SDK 获取并回写缓存。
// Redis 不可用时自动降级为 SDK 进程内缓存，不影响调用方。
func GetAccessToken() (string, error) {
	clientKey := variable.ConfigYml.GetString("Douyin.ClientKey")
	cacheKey := fmt.Sprintf("douyin:access_token:%s", clientKey)

	// 1. 优先读 Redis 缓存
	if redCli := redis_factory.GetOneRedisClient(); redCli != nil {
		token, err := redCli.String(redCli.Execute("GET", cacheKey))
		redCli.ReleaseOneRedisClient()
		if err == nil && token != "" {
			return token, nil
		}
	}

	// 2. 缓存未命中/Redis 异常，走 SDK 获取（注意用 GetClientToken，开放平台应用用
	// GetAccessToken 会打小程序端点报 40015 bad appid）
	cli, err := newSDKClient()
	if err != nil {
		return "", errors.Join(err, fmt.Errorf("douyin NewClient"))
	}
	token, err := cli.Credential.GetClientToken()
	if err != nil {
		return "", errors.Join(err, fmt.Errorf("douyin GetClientToken"))
	}
	accessToken := tea.StringValue(token.AccessToken)
	if accessToken == "" {
		return "", fmt.Errorf("douyin access_token 为空")
	}

	// 3. 回写 Redis，提前 5 分钟过期避免使用临界过期 token（尽力而为，失败不影响返回）
	ttl := int(tea.Int64Value(token.ExpiresIn)) - 300
	if ttl > 0 {
		if redCli := redis_factory.GetOneRedisClient(); redCli != nil {
			if _, err := redCli.Execute("SET", cacheKey, accessToken, "EX", ttl); err != nil {
				variable.ZapLog.Warn("douyin access_token 回写 Redis 失败", zap.Error(err))
			}
			redCli.ReleaseOneRedisClient()
		}
	}
	return accessToken, nil
}
