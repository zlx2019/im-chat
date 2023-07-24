package main

import (
	"regexp"
	"strings"
)

const (
	// 公开消息,所有人可见
	Public Type = iota
	// 私聊消息,仅对方可见
	Private
	// 系统消息,所有人不可见
	Admin
)

const (
	// 私聊报文协议正则匹配格式: `@[用户名] [消息]`
	// 如 `@lisi 你好啊` `@zhangsan 嗯嗯你好`
	PrivatePatter = "^@\\w* +.*$"
)

// message type
// 消息类型
type Type int8

// the user send Message of protocal
type Message struct {
	// 消息发送者
	User *User
	// 消息内容
	Payload string
	// 消息类型
	Type Type
	// 私聊目标用户名
	Target string
}

// new message protocol
// 创建消息协议
func OfMessage(user *User, payload string) *Message {
	return &Message{
		User:    user,
		Payload: payload,
	}
}

// 创建消息
func NewMessage(user *User, payload string, t Type) *Message {
	return &Message{
		User:    user,
		Payload: payload,
		Type:    t,
	}
}

// clone new message
// 根据消息类型,复制一份新的消息
func (self *Message) Clone(t Type) *Message {
	return &Message{
		User:    self.User,
		Payload: self.Payload,
		Type:    t,
	}
}

// 判断消息类型 是公开消息|私聊消息|操作命令等
func (self *Message) Parse() *Message {
	msg := self.Payload
	// 正则匹配消息格式
	if isPrivate, _ := regexp.MatchString(PrivatePatter, msg); isPrivate {
		// 私聊消息,解析消息报文
		self.UnPack()
	} else {
		// 公开消息 or 操作命令
		self.Type = Public
	}
	return self
}

// 对私聊消息进行解包, 获取私聊者和消息内容
func (self *Message) UnPack() {
	// 获取私聊用户名
	target := self.Payload[1:strings.Index(self.Payload, " ")]
	// 获取消息内容
	message := self.Payload[strings.Index(self.Payload, " "):]
	self.Payload = strings.TrimSpace(message)
	self.Target = target
	self.Type = Private
}
