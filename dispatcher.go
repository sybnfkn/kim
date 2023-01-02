package kim

import "github.com/klintcheng/kim/wire/pkt"

// Dispatcher defined a component how a message be dispatched to gateway
// 一个消息被分发到网关
type Dispatcher interface {
	Push(gateway string, channels []string, p *pkt.LogicPkt) error
}
