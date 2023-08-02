package client

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// 服务端地址信息,通过命令行传递
var DefaultServerIP string
var DefaultServerPort int
var DefaultName string

// 初始化命令行解析
func init() {
	flag.StringVar(&DefaultServerIP, "h", "127.0.0.1", "服务端IP")
	flag.IntVar(&DefaultServerPort, "p", 7080, "服务端口")
	flag.StringVar(&DefaultName, "n", "", "聊天用户名")
}

// 客户端
type Client struct {
	// 服务端IP
	ServerIP string
	// 服务端端口
	ServerPort int
	// 服务端连接
	conn net.Conn

	// 用户名
	Name string
	// 关闭客户端信号
	Stop chan struct{}
}

// 创建客户端
func NewClient() *Client {
	client := &Client{
		ServerIP:   DefaultServerIP,
		ServerPort: DefaultServerPort,
		Name:       DefaultName,
		Stop:       make(chan struct{}),
	}
	// 连接服务端
	client.Connect()
	return client
}

// 连接服务端
func (c *Client) Connect() {
	// 连接服务端
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.ServerIP, c.ServerPort))
	if err != nil {
		log.Printf("connect server failed: %s \n", err)
		panic(err)
	}
	c.conn = conn
	log.Println("connect server successfully~")
	// 重命名
	if c.Name != "" {
		c.Rename()
	}
}

// Client Run
// 开始运行
func (c *Client) Send() {
	// 创建reader流,读取终端输入
	reader := bufio.NewReader(os.Stdout)
	for {
		// 阻塞读取终端输入,以\n换行符为结尾
		line, _ := reader.ReadString('\n')
		// 退出客户端
		if line == "exit\n" {
			c.Stop <- struct{}{}
			return
		}
		if len(strings.TrimSuffix(line, "\n")) == 0 {
			continue
		}
		// 发送消息
		_, err := c.conn.Write([]byte(line))
		if err != nil {
			log.Printf("send message failed: %s \n", err.Error())
		}
	}
}

// 接收服务端数据,输出到终端
func (c *Client) Recv() {
	// 方式一
	// 将conn中的数据流 全部写入到Stdout中
	io.Copy(os.Stdout, c.conn)
	// 服务端连接断开,被动下线
	// close(c.Stop)
	c.Stop <- struct{}{}
}

// 客户端重命名
func (c *Client) Rename() {
	c.conn.Write([]byte("rename " + c.Name + "\n"))
}

// 关闭客户端
func (c *Client) Close() error {
	return c.conn.Close()
}
