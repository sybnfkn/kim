package pkt

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/klintcheng/kim/wire"
)

// **** 逻辑协议与基础协议都实现了Packet接口**** ，它有两个方法，Decode反序列化和Encode序列化，
// 逻辑协议中的Header和Body使用protobuf序列化框架可以减少手动编写Decode和Encode的代码，而基础协议则使用小头字节序手动处理
type Packet interface {
	Decode(r io.Reader) error
	Encode(w io.Writer) error
}

func MustReadLogicPkt(r io.Reader) (*LogicPkt, error) {
	val, err := Read(r)
	if err != nil {
		return nil, err
	}
	if lp, ok := val.(*LogicPkt); ok {
		return lp, nil
	}
	return nil, fmt.Errorf("packet is not a logic packet")
}

func MustReadBasicPkt(r io.Reader) (*BasicPkt, error) {
	val, err := Read(r)
	if err != nil {
		return nil, err
	}
	if bp, ok := val.(*BasicPkt); ok {
		return bp, nil
	}
	return nil, fmt.Errorf("packet is not a basic packet")
}

// 指定两个不同的魔数，就可以在网关层区分是基础协议还是逻辑协议
func Read(r io.Reader) (interface{}, error) {
	magic := wire.Magic{}
	_, err := io.ReadFull(r, magic[:])
	if err != nil {
		return nil, err
	}

	switch magic {
	case wire.MagicLogicPkt: // 逻辑logic
		p := new(LogicPkt)
		if err := p.Decode(r); err != nil {
			return nil, err
		}
		return p, nil
	case wire.MagicBasicPkt: // 基础logic
		p := new(BasicPkt)
		if err := p.Decode(r); err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, fmt.Errorf("magic code %s is incorrect", magic)
	}
}

// 一个封包的方法，主要作用是把Magic封装到消息的头部
func Marshal(p Packet) []byte {
	buf := new(bytes.Buffer)
	kind := reflect.TypeOf(p).Elem()

	// 反射处理
	if kind.AssignableTo(reflect.TypeOf(LogicPkt{})) {
		_, _ = buf.Write(wire.MagicLogicPkt[:])
	} else if kind.AssignableTo(reflect.TypeOf(BasicPkt{})) {
		_, _ = buf.Write(wire.MagicBasicPkt[:])
	}
	_ = p.Encode(buf)
	return buf.Bytes()
}
