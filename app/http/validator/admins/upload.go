package admins

import (
	"ofdhq-api/app/http/controller/admin"

	"github.com/gin-gonic/gin"
)

type UploadFile struct {
}

func (u UploadFile) CheckParams(context *gin.Context) {
	(&admin.Upload{}).Upload(context)
}
