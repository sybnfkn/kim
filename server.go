package kim

import (
	"context"
	"net"
	"time"
)

const (
	DefaultReadWait  = time.Minute * 3
	DefaultWriteWait = time.Second * 10
	DefaultLoginWait = time.Second * 10
	DefaultHeartbeat = time.Second * 55
)

const (
	// 定义读取消息的默认goroutine池大小
	DefaultMessageReadPool = 5000
	DefaultConnectionPool  = 5000
)

// 定义了基础服务的抽象接口
type Service interface {
	ServiceID() string
	ServiceName() string
	GetMeta() map[string]string
}

// 定义服务注册的抽象接口
type ServiceRegistration interface {
	Service
	PublicAddress() string
	PublicPort() int
	DialURL() string
	GetTags() []string
	GetProtocol() string
	GetNamespace() string
	String() string
}

// Server 定义了一个tcp/websocket不同协议通用的服务端的接口
type Server interface {
	ServiceRegistration
	// SetAcceptor 设置Acceptor
	SetAcceptor(Acceptor)
	//SetMessageListener 设置上行消息监听器
	SetMessageListener(MessageListener)
	//SetStateListener 设置连接状态监听服务,将连接断开的事件上报给业务层，让业务层可以实现一些逻辑处理。
	SetStateListener(StateListener)
	// SetReadWait 设置读超时,用于控制心跳逻辑
	SetReadWait(time.Duration)
	// SetChannelMap 设置Channel管理服务,设置一个连接管理器，Server在内部会自动管理连接的生命周期
	SetChannelMap(ChannelMap)

	// Start 用于在内部实现网络端口的监听和接收连接，
	// 并完成一个Channel的初始化过程。
	Start() error
	// Push 消息到指定的Channel中
	//  string channelID
	//  []byte 序列化之后的消息数据
	Push(string, []byte) error
	// Shutdown 服务下线，关闭连接
	Shutdown(context.Context) error
}

// Acceptor 连接接收器
type Acceptor interface {
	// Accept 返回一个握手完成的Channel对象或者一个error。
	// 业务层需要处理不同协议和网络环境下的连接握手协议
	Accept(Conn, time.Duration) (string, Meta, error)
}

// MessageListener 监听消息
type MessageListener interface {
	// 收到消息回调, Receive方法中第一个参数Agent表示发送方
	Receive(Agent, []byte)
}

// StateListener 状态监听器
type StateListener interface {
	// 连接断开回调
	Disconnect(string) error
}

type Meta map[string]string

// Agent is interface of client side
type Agent interface {
	ID() string        // ID : 返回连接的channelID。
	Push([]byte) error // Push : 用于上层业务返回消息
	GetMeta() Meta
}

// Conn Connection
// 通过对net.Conn进行二次包装，把读与写的操作封装到连接中，因此我们定义一个kim.Conn接口，继承了net.Conn接口。
type Conn interface {
	net.Conn
	// ReadFrame 与 WriteFrame，完成对websocket/tcp两种协议的封包与拆包逻辑的包装。
	// 读一个完整的协议包
	ReadFrame() (Frame, error)
	// 写一个完整的协议包
	WriteFrame(OpCode, []byte) error
	Flush() error
}

// Channel is interface of client side
type Channel interface {
	Conn // kim.Conn
	Agent
	// Close 关闭连接,重写net.Conn中的Close方法
	Close() error
	Readloop(lst MessageListener) error
	// SetWriteWait 设置写超时
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
}

// Client is interface of client side
type Client interface {
	Service
	// connect to server, 主动向一个服务器地址发起连接
	Connect(string) error
	// SetDialer 设置拨号处理器,设置一个拨号器，这个方法会在Connect中被调用，完成连接的建立和握手。
	SetDialer(Dialer)
	Send([]byte) error
	Read() (Frame, error)
	// Close 关闭
	Close()
}

// Dialer Dialer
//Send：发送消息到服务端。
//Read：读取一帧数据，这里底层复用了kim.Conn，所以直接返回Frame。
//Close：断开连接，退出。
type Dialer interface {
	DialAndHandshake(DialerContext) (net.Conn, error)
}

type DialerContext struct {
	Id      string
	Name    string
	Address string
	Timeout time.Duration
}

// OpCode OpCode
type OpCode byte

// Opcode type
// 从Websocket协议抄过来的
const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

// Frame Frame
// TCP协议是流式传输，通常需要上层业务处理拆包。
//Websocket协议是基于Frame，在底层Server中就可以区分出每一个Frame，然后把Frame中的Payload交给上层。
//通过抽象一个Frame接口来解决底层封包与拆包问题。
type Frame interface {
	SetOpCode(OpCode) // OpCode表示操作类型
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}
