package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
)

// 用户信息
type User struct {
	// 用户名
	Name string
	// 用户地址信息
	Addr string
	// 用户接收消息通道
	Ch chan string
	// 用户的客户端连接
	conn net.Conn
	// 会话上下文
	StopCtx context.Context
	// 结束会话触发方法
	StopCancel context.CancelFunc
}

// NewUser new user info
// `Name` the default client address
// 创建一个上线用户
func NewUser(conn net.Conn) *User {
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
func (u *User) Reader(s *Server) {
	buf := make([]byte, 1024)
	for {
		n, err := u.conn.Read(buf)
		fmt.Println(n, err)
		// 读取数据错误
		if err != nil {
			// 用户客户端关闭
			if err == io.EOF {
				// 广播下线消息
				s.Pushlish(u, "我下线咯~")
				// 加锁后 将用户踢出在线列表
				s.Lock.Lock()
				delete(s.OnlineUsers, u.Name)
				s.Lock.Unlock()
				// 关闭消息通道、停止用户写协程
				close(u.Ch)
				u.StopCancel()
				return
			}
			log.Println("client read error:", err)
			continue
		}
		// 读取数据成功(去除 '\n')
		message := string(buf[:n-1])
		s.Pushlish(u, message)
	}
}