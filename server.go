package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
)

// 服务端
type Server struct {
	// 服务端IP
	IP string
	// 服务端口
	Port int
	// 在线用户列表 k: 用户名 v: 用户信息
	OnlineUsers map[string]*User
	// 锁
	Lock sync.RWMutex
	// 消息广播器,所有的消息都从这里广播出去
	Publisher chan string
}

// 创建服务端
// new server
func NewServer(ip string, port int) *Server {
	return &Server{
		IP:          ip,
		Port:        port,
		OnlineUsers: make(map[string]*User),
		Publisher:   make(chan string, 10),
	}
}

// run server
// 服务器开始运行
func (s *Server) Run() {
	// socket listen
	// 创建TCP服务
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		log.Println("net server run error")
		panic(err)
	}

	// close socket listen
	// 关闭网络服务
	defer listen.Close()

	// listener message
	go s.Listener()

	log.Println("聊天室已启动~")
	log.Println("当前协程数量: ", runtime.NumGoroutine())
	// loop do business
	// 循环处理业务
	for {
		// accept client conn
		// 等待客户端连接
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("client conn error: %s \n", err.Error())
			continue
		}
		// do handler
		// 处理客户端连接 用户上线
		go s.Online(conn)
	}
}

// 处理客户端连接,用户上线
func (s *Server) Online(conn net.Conn) {
	log.Printf("client conn in addr: %s \n", conn.LocalAddr().String())
	// 创建用户
	user := NewUser(conn)
	// 添加至在线用户列表
	s.Lock.Lock()
	s.OnlineUsers[user.Name] = user
	s.Lock.Unlock()
	onlineMsg := user.Addr + " 上线辣~"
	// 广播用户上线消息
	s.Pushlish(user, onlineMsg)
	// 开启一个协程 读取服务端消息,并且写回客户端
	go user.Writer()
	// 开启一个协程 读取客户端消息,发送给服务端广播器
	go user.Reader(s)
	log.Println(onlineMsg)
	log.Println("当前协程数量: ", runtime.NumGoroutine())
}

// 将消息推送给广播器
func (s *Server) Pushlish(user *User, msg string) {
	s.Publisher <- fmt.Sprintf("[%s]: %s", user.Name, msg)
}

// 消息监听,并且广播给所有在线用户
func (s *Server) Listener() {
	for {
		message, ok := <-s.Publisher
		if !ok {
			// 广播器已经关闭
			return
		}
		// 广播消息
		s.Lock.RLock()
		for _, u := range s.OnlineUsers {
			u.Ch <- message
		}
		s.Lock.RUnlock()
	}
}
