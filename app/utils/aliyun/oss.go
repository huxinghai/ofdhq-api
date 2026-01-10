package aliyun

import (
	"errors"
	"fmt"
	"mime/multipart"
	"ofdhq-api/app/global/variable"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
)

func UploadFileOSS(uploadFile *multipart.FileHeader, bucketType int32) (string, error) {
	if uploadFile == nil {
		return "", fmt.Errorf("uploadFile is nil")
	}

	file, err := uploadFile.Open()
	if err != nil {
		return "", errors.Join(err, fmt.Errorf("uploadfile.Open"))
	}
	defer file.Close()

	curYearMonth := time.Now().Format("2006_01")

	newUUID := uuid.New()

	key := newUUID.String()
	if len(key) > 10 {
		key = key[:10]
	}

	path := "assets/" + curYearMonth + "/" + key + uploadFile.Filename

	host := variable.ConfigYml.GetString("Aliyun.host")
	accessKeyID := variable.ConfigYml.GetString("Aliyun.access_key_id")
	accessKeySecret := variable.ConfigYml.GetString("Aliyun.access_key_secret")
	bucketName := variable.ConfigYml.GetString("Aliyun.bucket")
	if bucketType == 1 {
		bucketName = variable.ConfigYml.GetString("Aliyun.bucketweb")
		host = variable.ConfigYml.GetString("Aliyun.web_host")
	}

	client, err := oss.New("http://oss-cn-shenzhen.aliyuncs.com", accessKeyID, accessKeySecret)
	if err != nil {
		return "", fmt.Errorf("aliyun oss.New err:%v", err)
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", fmt.Errorf("client.Bucket err:%v", err)
	}

	err = bucket.PutObject(path, file)
	if err != nil {
		return "", fmt.Errorf("bucket.PutObjectFromFile err:%v, path:%+v", err, path)
	}
	return host + "/" + path, nil
}
