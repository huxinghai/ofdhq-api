package test

import (
	"fmt"
	"ofdhq-api/app/global/variable"
	_ "ofdhq-api/bootstrap" //  为了保证单元测试与正常启动效果一致，记得引入该包
	"testing"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
	"github.com/qifengzhang007/goCurl"
)

// goCurl 更详细的使用文档 https://gitee.com/daitougege/goCurl

// 一个简单的get请求
func TestHttpClient(t *testing.T) {
	cli := goCurl.CreateHttpClient()
	if resp, err := cli.Get("http://hq.sinajs.cn/list=sh601360"); err == nil {
		content, err := resp.GetContents()
		if err != nil {
			t.Errorf("单元测试未通过,返回值不符合要求：%s\n", content)
		}
		t.Log(content)
	}
}

// 向门户服务接口请求，用于收集cpu占用情况。
func TestPprof(t *testing.T) {
	cli := goCurl.CreateHttpClient()
	for i := 1; i <= 500; i++ {
		resp, err := cli.Get("http://127.0.0.1:20191/api/v1/home/news", goCurl.Options{
			FormParams: map[string]interface{}{
				"newsType": "portal",
				"page":     "2",
				"limit":    "52",
			},
		})
		if err == nil {
			if txt, err := resp.GetContents(); err == nil {
				if i == 500 {
					//最后一次输出返回结果，避免中间过程频繁操作io
					variable.ZapLog.Info(txt)
				}
			}
		} else {
			t.Log(err.Error())
		}
	}
}

func TestUploadFile(t *testing.T) {
	curYearMonth := time.Now().Format("2006_01")

	newUUID := uuid.New()

	key := newUUID.String()
	if len(key) > 10 {
		key = key[:10]
	}

	path := "assets/" + curYearMonth + "/" + key + "test.jpg"

	accessKeyID := variable.ConfigYml.GetString("Aliyun.access_key_id")
	accessKeySecret := variable.ConfigYml.GetString("Aliyun.access_key_secret")
	bucketName := variable.ConfigYml.GetString("Aliyun.bucketweb")

	// host := variable.ConfigYml.GetString("Aliyun.host")

	client, err := oss.New("http://oss-cn-shenzhen.aliyuncs.com", accessKeyID, accessKeySecret)
	if err != nil {
		t.Errorf("aliyun oss.New err:%v", err)
		return
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		t.Errorf("client.Bucket err:%v", err)
		return
	}

	err = bucket.PutObjectFromFile(path, "/Users/edy/Downloads/unnamed.jpg")
	if err != nil {
		t.Errorf("bucket.PutObjectFromFile err:%v, path:%+v", err, path)
		return
	}

	fmt.Printf("完成 \n")
}
