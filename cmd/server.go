package main

import "chat/server"

// 服务端编译入口
func main() {
	// 创建服务
	s := server.NewServer("127.0.0.1", 7080)
	// 启动
	s.Run()
}
