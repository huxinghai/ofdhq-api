package api

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/utils/aliyun"
	"ofdhq-api/app/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Upload struct {
}

const maxSize = 1 * 1024 * 1024 // 1MB

func (t *Upload) Upload(ctx *gin.Context) {
	tmpFile, err := ctx.FormFile("file")
	if err != nil {
		response.Fail(ctx, consts.FilesUploadFailCode, consts.FilesUploadFailMsg, err.Error())
		return
	}

	if tmpFile == nil {
		response.Fail(ctx, consts.FilesUploadFailCode, "file 参数不能为空", nil)
		return
	}

	if tmpFile.Size > maxSize {
		response.Fail(ctx, consts.FilesUploadFailCode, "文件过大，最大允许 1MB", nil)
		return
	}

	urlPath, err := aliyun.UploadFileOSS(tmpFile)
	if err != nil {
		variable.ZapLog.Error("aliyun.UploadFileOSS 失败", zap.Error(err))
		response.Fail(ctx, consts.FilesUploadFailCode, "上传文件失败", nil)
		return
	}

	res := gin.H{
		"url": urlPath,
	}
	response.Success(ctx, consts.CurdStatusOkMsg, res)
}
