package gochannel

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/nats-io/nats.go"
)

type Channel[T any] struct {
	id       string //这是channcel的唯一ID，主要用于进行订阅使用
	subject  string
	natsConn *nats.Conn
}

func NewChannel[T any](optFn ...func(*Option)) *Channel[T] {
	var opt Option

	if len(optFn) > 0 {
		for _, fn := range optFn {
			fn(&opt)
		}
	}

	c := &Channel[T]{
		id: guid.S(),
	}
	c.subject = fmt.Sprintf("Chan_%s", c.id)

	if opt.NatsConn != nil {
		c.natsConn = opt.NatsConn
	} else {
		var err error
		c.natsConn, err = nats.Connect(nats.DefaultURL)
		if err != nil {
			g.Log().Error(context.Background(), err)
		}
	}

	return c
}

func (c *Channel[T]) Send(ctx context.Context, data T) {
	if bs, err := gjson.Encode(data); err == nil {
		c.natsConn.Publish(c.subject, bs)
	}
}

func (c *Channel[T]) SendChan(ctx context.Context) (sendCh chan<- T, closeFn func()) {
	ch := make(chan T, 100)
	sendCh = ch
	closeFn = func() {
		close(ch)
	}
	go func() {
		defer close(ch)
		for {
			select {
			case data, ok := <-ch:
				if !ok {
					return
				}
				c.Send(ctx, data)
			case <-ctx.Done():
				return
			}
		}
	}()
	return
}

func (c *Channel[T]) Recv(ctx context.Context) (data T, ok bool) {
	//等待接收，需要通知网络上有消费者
	sub, err := c.natsConn.QueueSubscribeSync(c.subject, "default")
	if err != nil {
		ok = false
		return
	}
	defer sub.Unsubscribe()
	msg, err := sub.NextMsgWithContext(ctx)
	if err != nil {
		ok = false
		return
	}
	if err := gjson.DecodeTo(msg.Data, &data); err == nil {
		return
	}

	return
}

func (c *Channel[T]) RecvChan(ctx context.Context) (recvChan <-chan T, err error) {
	//等待接收，需要通知网络上有消费者
	data := make(chan T, 100)
	cm := make(chan *nats.Msg, 100)
	sub, err := c.natsConn.QueueSubscribeSyncWithChan(c.subject, "default", cm)
	if err != nil {
		return
	}

	go func() {
		defer close(data)
		defer close(cm)
		defer sub.Unsubscribe()

		for {
			select {
			case msg, ok := <-cm:
				if !ok {

					return
				}
				var one T
				if err := gjson.DecodeTo(msg.Data, &one); err == nil {
					data <- one
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return data, nil
}
