package redis_factory

import (
	"errors"
	"fmt"
	"ofdhq-api/app/global/my_errors"
	"ofdhq-api/app/global/variable"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DuplicateFactory struct {
	redisClient *RedisClient
	key         string
	seconds     int
	value       string
}

func NewDuplicateFactory(key string, seconds int) *DuplicateFactory {
	redCli := GetOneRedisClient()
	newUUID := uuid.New()
	return &DuplicateFactory{redisClient: redCli, key: fmt.Sprintf("duplicate:key:%s", key), seconds: seconds, value: newUUID.String()}
}

func (d *DuplicateFactory) IsDuplicateRequest() error {

	// 使用 SetNX 命令尝试将 key 设置到 Redis 中
	result, err := d.redisClient.String(d.redisClient.Execute("SET", d.key, d.value, "NX", "EX", d.seconds))
	if err != nil {
		return errors.Join(err, fmt.Errorf("执行失败！key:%s", d.key))
	}
	if strings.ToUpper(result) != "OK" {
		return my_errors.ErrDuplicateRequest
	}
	return nil
}

func (d *DuplicateFactory) Clean() error {
	v, err := d.redisClient.Int64(d.redisClient.Execute("EVAL", `
	if redis.call("get",KEYS[1]) == ARGV[1]
	then
		return redis.call("del",KEYS[1])
	else
		return 0
	end`, 1, d.key, d.value))

	if err != nil {
		variable.ZapLog.Error("执行脚本失败", zap.Error(err))
		return errors.Join(err, fmt.Errorf("处理失败！d:%+v", d))
	}
	if v != 1 {
		return fmt.Errorf("解锁处理失败！d:%+v", d)
	}
	return nil
}
