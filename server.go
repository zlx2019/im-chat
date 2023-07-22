package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
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
	Publisher chan *Message
}

// 创建服务端
// new server
func NewServer(ip string, port int) *Server {
	return &Server{
		IP:          ip,
		Port:        port,
		OnlineUsers: make(map[string]*User),
		Publisher:   make(chan *Message, 10),
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
		go s.Handler(conn)
	}
}

// 处理客户端连接,用户上线
func (s *Server) Handler(conn net.Conn) {
	// log.Printf("client conn in addr: %s \n", conn.LocalAddr().String())
	// 创建用户
	user := NewUser(conn, s)
	// 用户上线
	user.Online()
	log.Println("当前协程数量: ", runtime.NumGoroutine())

	for {
		select {
		case <-user.Active:
			// 接收心跳信号
		case <-time.After(time.Second * 10):
			// 一旦用户超过10秒没有收到心跳信号,超时强制下线
			user.conn.Write([]byte("已超时,您被强制下线了~"))
			s.Pushlish(OfMessage(user, "已超时,强制下线~", Admin))
			// 用户下线
			user.Downline()
			// 退出当前协程
			return
		case <-user.StopCtx.Done():
			return
		}
	}

}

// 将消息推送给广播器
func (s *Server) Pushlish(msg *Message) {
	s.Publisher <- msg
}

// 消息监听,并且广播给所有在线用户
func (s *Server) Listener() {
	for {
		m, ok := <-s.Publisher
		if !ok {
			// 广播器已经关闭
			return
		}

		// 系统消息
		if m.Type == Admin {
			log.Printf("[%s] %s", m.User.Name, m.Payload)
		}

		// 广播消息
		for _, u := range s.OnlineUsers {
			if m.Type == Public {
				// 公开消息,排除掉发送者本身
				if u.Name == m.User.Name {
					continue
				}
				u.Ch <- fmt.Sprintf("[%s]: %s", m.User.Name, m.Payload)
			} else if m.Type == Private {
				// 私信消息
				// TODO
			}
		}
	}
}
