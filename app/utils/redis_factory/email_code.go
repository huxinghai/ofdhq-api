package redis_factory

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type emailCode struct {
	key         string
	email       string
	redisClient *RedisClient
}

func NewEmailCode(email string) *emailCode {
	key := fmt.Sprintf("code:email:%s", email)
	redCli := GetOneRedisClient()
	return &emailCode{
		email:       email,
		redisClient: redCli,
		key:         key,
	}
}

func (f *emailCode) GenCode() (string, error) {
	code := generateRandomCode()
	codestr := strconv.FormatInt(int64(code), 10)

	_, err := f.redisClient.Execute("SET", f.key, codestr, "EX", 1800)
	if err != nil {
		return "", errors.Join(err, fmt.Errorf("生成Code 失败！"))
	}

	return codestr, nil
}

func (f *emailCode) ValidCode(code string) (bool, error) {
	codeT, err := f.redisClient.String(f.redisClient.Execute("GET", f.key))
	if err != nil {
		return false, errors.Join(err, fmt.Errorf("验证Code 失败！"))
	}

	if code == codeT {
		return true, nil
	}

	return false, nil
}

func generateRandomCode() int {
	source := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(source)
	return randomGenerator.Intn(90000) + 10000
}
