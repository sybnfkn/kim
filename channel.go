package kim

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/klintcheng/kim/logger"
	"github.com/panjf2000/ants/v2"
)

// ChannelImpl is a websocket implement of channel
type ChannelImpl struct {
	id string
	Conn
	meta      Meta
	writechan chan []byte //writechan的管道，发送的消息直接通过writechan发送给了一个独立的goruntine中writeloop()执行，这样就使得Push变成了一个线程安全方法。
	writeWait time.Duration
	readwait  time.Duration
	gpool     *ants.Pool
	state     int32 // 0 init 1 start 2 closed
}

// NewChannel NewChannel
func NewChannel(id string, meta Meta, conn Conn, gpool *ants.Pool) Channel {
	ch := &ChannelImpl{
		id:        id,
		Conn:      conn,
		meta:      meta,
		writechan: make(chan []byte, 5),
		writeWait: DefaultWriteWait, //default value
		readwait:  DefaultReadWait,
		gpool:     gpool,
		state:     0,
	}
	go func() {
		err := ch.writeloop()
		if err != nil {
			logger.WithFields(logger.Fields{
				"module": "ChannelImpl",
				"id":     id,
			}).Info(err)
		}
	}()
	return ch
}

// 一个独立的goruntine中writeloop()执行
func (ch *ChannelImpl) writeloop() error {
	log := logger.WithFields(logger.Fields{
		"module": "ChannelImpl",
		"func":   "writeloop",
		"id":     ch.id,
	})
	defer func() {
		log.Debugf("channel %s writeloop exited", ch.id)
	}()
	for payload := range ch.writechan {
		err := ch.WriteFrame(OpBinary, payload)
		if err != nil {
			return err
		}
		chanlen := len(ch.writechan)
		for i := 0; i < chanlen; i++ {
			payload = <-ch.writechan
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}
		}
		err = ch.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

// ID id simpling server
func (ch *ChannelImpl) ID() string { return ch.id }

// Push 异步写数据
func (ch *ChannelImpl) Push(payload []byte) error {
	if atomic.LoadInt32(&ch.state) != 1 {
		return fmt.Errorf("channel %s has closed", ch.id)
	}
	// 异步写, 交给一个writeloop()执行， 在chan的读取那边
	ch.writechan <- payload
	return nil
}

// Close 关闭连接
func (ch *ChannelImpl) Close() error {
	if !atomic.CompareAndSwapInt32(&ch.state, 1, 2) {
		return fmt.Errorf("channel has started")
	}
	close(ch.writechan)
	return nil
}

// SetWriteWait 设置写超时
func (ch *ChannelImpl) SetWriteWait(writeWait time.Duration) {
	if writeWait == 0 {
		return
	}
	ch.writeWait = writeWait
}

func (ch *ChannelImpl) SetReadWait(readwait time.Duration) {
	if readwait == 0 {
		return
	}
	ch.writeWait = readwait
}

// Readloop 把消息的读取和心跳处理的逻辑封装在一起。
func (ch *ChannelImpl) Readloop(lst MessageListener) error {
	if !atomic.CompareAndSwapInt32(&ch.state, 0, 1) {
		return fmt.Errorf("channel has started")
	}
	log := logger.WithFields(logger.Fields{
		"struct": "ChannelImpl",
		"func":   "Readloop",
		"id":     ch.id,
	})
	for {
		_ = ch.SetReadDeadline(time.Now().Add(ch.readwait))

		// kim.Conn中的ReadFrame()
		frame, err := ch.ReadFrame()
		if err != nil {
			log.Info(err)
			return err
		}
		if frame.GetOpCode() == OpClose {
			return errors.New("remote side close the channel")
		}
		// 处理ping
		if frame.GetOpCode() == OpPing {
			log.Trace("recv a ping; resp with a pong")

			_ = ch.WriteFrame(OpPong, nil)
			_ = ch.Flush()
			continue
		}
		payload := frame.GetPayload()
		if len(payload) == 0 {
			continue
		}
		err = ch.gpool.Submit(func() {
			//MessageListener 用作消息处理
			lst.Receive(ch, payload)
		})
		if err != nil {
			return err
		}
	}
}

func (ch *ChannelImpl) GetMeta() Meta { return ch.meta }
