package routers

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/http/middleware/cors"
	validatorFactory "ofdhq-api/app/http/validator/core/factory"
	"ofdhq-api/app/utils/gin_release"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func InitApiRouter() *gin.Engine {
	var router *gin.Engine
	// 非调试模式（生产模式） 日志写到日志文件
	if variable.ConfigYml.GetBool("AppDebug") == false {
		//【生产模式】
		// 根据 gin 官方的说明：[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
		// 如果部署到生产环境，请使用以下模式：
		// 1.生产模式(release) 和开发模式的变化主要是禁用 gin 记录接口访问日志，
		// 2.go服务就必须使用nginx作为前置代理服务，这样也方便实现负载均衡
		// 3.如果程序发生 panic 等异常使用自定义的 panic 恢复中间件拦截、记录到日志
		router = gin_release.ReleaseRouter()
	} else {
		// 调试模式，开启 pprof 包，便于开发阶段分析程序性能
		router = gin.Default()
		pprof.Register(router)
	}

	//根据配置进行设置跨域
	if variable.ConfigYml.GetBool("HttpServer.AllowCrossDomain") {
		router.Use(cors.Next())
	}

	vApi := router.Group("/api/v1/")
	{
		topicNotNeedAuth := vApi.Group("topic/")
		{
			topicNotNeedAuth.GET("list", validatorFactory.Create(consts.ValidatorPrefix+"TopicList"))
			topicNotNeedAuth.GET("detail", validatorFactory.Create(consts.ValidatorPrefix+"TopicDetail"))
		}

		vApi.POST("customer/create", validatorFactory.Create(consts.ValidatorPrefix+"CustomerCreate"))
		vApi.POST("upload/file", validatorFactory.Create(consts.ValidatorPrefix+"UserUploadFile"))

		//----------------------- 需要登录态接口 ----------------------
		// vApi.Use(authorization.CheckTokenAuth())
	}

	InitAdminApiRouter(router)
	return router
}
