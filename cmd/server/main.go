package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mock-storage/internal/service"
)

func main() {
	// 创建服务实例
	storageService, err := service.NewObjectStorageService()
	if err != nil {
		fmt.Printf("创建服务失败: %v\n", err)
		os.Exit(1)
	}

	// 启动服务
	err = storageService.Start()
	if err != nil {
		fmt.Printf("启动服务失败: %v\n", err)
		os.Exit(1)
	}

	// 设置优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	println("服务正在运行...")
	<-quit

	// 停止服务
	err = storageService.Stop()
	if err != nil {
		fmt.Printf("停止服务时出错: %v\n", err)
	}
}
