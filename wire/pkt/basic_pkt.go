package pkt

import (
	"io"

	"github.com/klintcheng/kim/wire/endian"
)

// basic pkt code
const (
	CodePing = uint16(1)
	CodePong = uint16(2)
)

/**
需要在业务层协议中支持ping/pong协议。但是如果我们直接在逻辑协议基础上添加心跳指令不太合适，主要有两点：
1。心跳在网关层就返回，不转发给逻辑服务处理。
2。心跳包要尽量小，逻辑协议的Header太重。
*/
type BasicPkt struct {
	// 1：ping    2：pong
	Code   uint16
	Length uint16
	Body   []byte
}

func (p *BasicPkt) Decode(r io.Reader) error {
	var err error
	if p.Code, err = endian.ReadUint16(r); err != nil {
		return err
	}
	if p.Length, err = endian.ReadUint16(r); err != nil {
		return err
	}
	if p.Length > 0 {
		if p.Body, err = endian.ReadFixedBytes(int(p.Length), r); err != nil {
			return err
		}
	}
	return nil
}

func (p *BasicPkt) Encode(w io.Writer) error {
	if err := endian.WriteUint16(w, p.Code); err != nil {
		return err
	}
	if err := endian.WriteUint16(w, p.Length); err != nil {
		return err
	}
	if p.Length > 0 {
		if _, err := w.Write(p.Body); err != nil {
			return err
		}
	}
	return nil
}
