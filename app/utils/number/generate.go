package number

import (
	"errors"
	"fmt"
	"ofdhq-api/app/utils/redis_factory"
	"time"
)

type NumberRedis struct {
	redisClient       *redis_factory.RedisClient
	generateNumberKey string
}

func NewNumberRedis() *NumberRedis {
	redCli := redis_factory.GetOneRedisClient()
	if redCli == nil {
		return nil
	}

	return &NumberRedis{
		redisClient:       redCli,
		generateNumberKey: "utils:number:generate:max-num",
	}
}

func (nr *NumberRedis) GetMaxNum() (int64, error) {
	t, err := nr.redisClient.Int64(nr.redisClient.Execute("INCR", nr.generateNumberKey))
	if err != nil {
		return 0, errors.New(fmt.Sprintf("生成Number Redis获取数据失败！err:%s", err.Error()))
	}

	return t, nil
}

func GenerateNumber(bizCode string) (string, error) {
	t := NewNumberRedis()
	maxNum, err := t.GetMaxNum()
	if err != nil {
		return "", err
	}
	tt := serializeWithLeadingZeros(int(maxNum), 3)
	uid := fmt.Sprintf("%s%s%s", bizCode, time.Now().Format("06010204"), tt)

	return uid, nil
}

// 使用 fmt.Sprintf 进行格式化，%0Nd 表示使用零进行填充，N 为补零后的总位数
func serializeWithLeadingZeros(number, length int) string {
	formatString := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(formatString, number)
}
