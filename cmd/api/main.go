package main

import (
	"context"
	"log"
	"net/http"
	"ofdhq-api/app/global/variable"
	_ "ofdhq-api/bootstrap"
	"ofdhq-api/routers"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router := routers.InitApiRouter()

	// 创建 HTTP Server
	srv := &http.Server{
		Addr:    variable.ConfigYml.GetString("HttpServer.Api.Port"),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	robotallot.InitAutoUserStrategyExpire(ctx)
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	robotallot.InitAutoKeepAliveRobot(ctx)
	// }()

	// 等待中断信号以优雅地关闭服务器（设置5秒的超时时间）
	<-ctx.Done()
	stop()
	log.Println("Shutdown Server ...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 HTTP Server
	// 	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	wg.Wait()
	log.Println("Server exiting")
}
