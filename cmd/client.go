package main

import (
	"chat/client"
	"flag"
	"log"
)

// 客户端编译入口
func main() {
	// 解析命令行
	flag.Parse()
	// 创建客户端
	cli := client.NewClient()
	defer cli.Close()
	// 开启一个协程读取消息
	go cli.Recv()
	// 开启一个协程发送消息
	go cli.Send()

	// 阻塞等待关闭信号
	select {
	case <-cli.Stop:
		log.Println("Client Quitting~")
	}
}
