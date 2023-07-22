package main

// main
func main() {
	// 创建服务
	s := NewServer("127.0.0.1", 7080)
	// 启动
	s.Run()
}
