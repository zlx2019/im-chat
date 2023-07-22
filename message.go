package main

const (
	// 公开消息,所有人可见
	Public Type = iota
	// 私聊消息,仅对方可见
	Private
	// 系统消息,所有人不可见
	Admin
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
}

// new message protocol
// 创建消息协议
func OfMessage(user *User, payload string, t Type) *Message {
	return &Message{
		User:    user,
		Payload: payload,
		Type:    t,
	}
}
