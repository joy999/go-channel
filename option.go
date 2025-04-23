package gochannel

import "github.com/nats-io/nats.go"

type Option struct {
	Size     int
	NatsConn *nats.Conn
}

func WithSize(size int) func(*Option) {
	return func(o *Option) {
		o.Size = size
	}
}

func WithNatsConn(conn *nats.Conn) func(*Option) {
	return func(o *Option) {
		o.NatsConn = conn
	}
}
