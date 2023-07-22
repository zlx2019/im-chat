package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"strings"
)

// 用户信息
type User struct {
	// 用户名
	Name string
	// 用户地址信息
	Addr string
	// 用户接收消息通道
	Ch chan string
	// 会话上下文
	StopCtx context.Context
	// 结束会话触发方法
	StopCancel context.CancelFunc
	// 用户的客户端连接
	conn net.Conn
	// 服务端
	server *Server
}

// NewUser new user info
// `Name` the default client address
// 创建一个上线用户
func NewUser(conn net.Conn, s *Server) *User {
	// 创建一个控制用户的上下文
	stop, canl := context.WithCancel(context.Background())
	stop.Done()
	return &User{
		Name:       conn.RemoteAddr().String(),
		Addr:       conn.RemoteAddr().String(),
		Ch:         make(chan string),
		conn:       conn,
		StopCtx:    stop,
		StopCancel: canl,
		server:     s,
	}
}

// run user read message from server
// 为当前用户开启协程, 循环从消息通道内读取消息,然后响应给客户端
func (u *User) Writer() {
	for {
		select {
		// 读取消息
		case msg := <-u.Ch:
			// 写回客户端
			u.conn.Write([]byte(msg + "\r\n"))
		case <-u.StopCtx.Done():
			// 客户端关闭
			log.Println(u.Name + " 下线了~")
			log.Println("当前协程数量: ", runtime.NumGoroutine())
			return
		}
	}
}

// read client message send to server
// 读取客户端消息,发送给服务端广播器
func (u *User) Reader() {
	buf := make([]byte, 1024)
	for {
		n, err := u.conn.Read(buf)
		// 读取数据错误
		if err != nil {
			// 用户客户端关闭
			if err == io.EOF {
				// 用户下线
				u.Downline()
				return
			}
			log.Println("client read error:", err)
			continue
		}
		// 读取数据成功(去除 '\n')
		message := string(buf[:n-1])
		// 是否是操作命令
		if message == "ls" {
			// 查看所有在线用户
			message = u.OnlineUsers()
			u.Ch <- message
		} else if len(message) > 7 && message[:7] == "rename " {
			// 重命名
			u.Rename(message[7:])
		} else {
			// 将消息转发给服务端
			u.server.Pushlish(OfMessage(u, message, Public))
		}
	}
}

// user go live. start read and write
// 用户上线
func (u *User) Online() {
	// 添加到服务器在线列表
	u.server.Lock.Lock()
	u.server.OnlineUsers[u.Name] = u
	u.server.Lock.Unlock()
	// 广播上线消息
	u.server.Pushlish(OfMessage(u, "上线辣~", Public))
	// 开启一个协程 读取服务端消息,并且写回客户端
	go u.Writer()
	// 开启一个协程 读取客户端消息,发送给服务端广播器
	go u.Reader()

}

// user out close resources
// 用户下线,释放用户相关资源
func (u *User) Downline() {
	// 从服务器在线列表中移除
	u.server.Lock.Lock()
	delete(u.server.OnlineUsers, u.Name)
	u.server.Lock.Unlock()
	// 停止用户循环写的协程
	u.StopCancel()
	// 广播下线消息
	u.server.Pushlish(OfMessage(u, "下线辣~", Public))

}

// find all online users
// 查询所有在线用户信息(排除自己)
func (u *User) OnlineUsers() string {
	names := []string{}
	u.server.Lock.Lock()
	for k := range u.server.OnlineUsers {
		if k != u.Name {
			names = append(names, fmt.Sprintf("[%s]", k))
		}
	}
	title := "当前在线用户如下: \n"
	u.server.Lock.Unlock()
	switch len(names) {
	case 0:
		return "暂无任何在线用户~"
	case 1:
		return title + names[0]
	default:
		return title + strings.Join(names, "\n")
	}
}

// rename
// 用户重命名
func (u *User) Rename(newName string) {
	u.server.Lock.Lock()
	defer u.server.Lock.Unlock()
	// 判断名称是否已存在
	if _, ok := u.server.OnlineUsers[newName]; ok {
		u.Ch <- "该名称已存在,请尝试其他名称!"
		return
	}
	delete(u.server.OnlineUsers, u.Name)
	u.Name = newName
	u.server.OnlineUsers[u.Name] = u
}
