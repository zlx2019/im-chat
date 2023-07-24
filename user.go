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
	// 心跳活跃信号
	Active chan struct{}
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
		Active:     make(chan struct{}),
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
			// 接收到Reader协程 停止信号,关闭本协程
			close(u.Ch)
			log.Println("当前协程数量: ", runtime.NumGoroutine()-1)
			Quit()
		}
	}
}

// read client message send to server
// 读取客户端消息,发送给服务端广播器
func (u *User) Reader() {
	buf := make([]byte, 1024)
	for {
		// 读取消息
		n, err := u.conn.Read(buf)
		if err != nil {
			// 判断客户端是否已经关闭,或者用户已经超时被动关闭
			if err == io.EOF {
				// 客户端主动关闭,用户下线
				u.Downline()
			} else if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				// conn已被close,可能已经超时强制退出.直接结束该协程
				Quit()
			}
			log.Println("client read error:", err)
			continue
		}

		// 读取消息并且,解析
		data := string(buf[:n-1])
		// 判空
		if strings.TrimSpace(data) == "" {
			continue
		}

		message := OfMessage(u, data).Parse()
		switch message.Type {
		case Private:
			// 私聊消息
			// 获取目标用户
			targetUser, ok := u.server.OnlineUsers[message.Target]
			if !ok {
				// 目标用户不存在
				u.Ch <- "您私聊的用户不存在!"
				continue
			}
			// 向目标用户 发现私信
			targetUser.Ch <- fmt.Sprintf("[%s]: %s", u.Name, message.Payload)
		case Public, Admin:
			// 公开消息 或者 操作命令
			m := message.Payload
			if m == "ls" {
				// 查看所有在线用户信息
				u.Ch <- u.OnlineUsers()
			} else if len(m) > 7 && m[:7] == "rename " {
				// 用户重命名
				u.Rename(m[7:])
			} else {
				// 公开消息,直接发送给服务端
				u.server.Pushlish(message)
			}
		}
		// 发送心跳信号,保持活跃 否则会超时
		u.Active <- struct{}{}
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
	message := NewMessage(u, "上线辣~", Public)
	u.server.Pushlish(message)
	u.server.Pushlish(message.Clone(Admin))
	// 开启一个协程 读取服务端消息,并且写回客户端
	go u.Writer()
	// 开启一个协程 读取客户端消息,发送给服务端广播器
	go u.Reader()

}

// user out close resources
// 用户下线,释放用户相关资源
func (u *User) Downline() {
	//1. 从服务器在线列表中移除
	u.server.Lock.Lock()
	delete(u.server.OnlineUsers, u.Name)
	u.server.Lock.Unlock()

	//2. 通过上下文结束用户Writer任务协程
	u.StopCancel()

	//3. 广播下线消息
	message := NewMessage(u, "下线辣~", Public)
	u.server.Pushlish(message)
	u.server.Pushlish(message.Clone(Admin))
	//4. 关闭客户端连接
	u.conn.Close()

	// 5. 退出当前协程
	Quit()
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

// 结束当前协程
func Quit() {
	runtime.Goexit()
}
