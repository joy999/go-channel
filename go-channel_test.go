package gochannel_test

import (
	"context"
	"fmt"
	"gochannel"
	"testing"
	"time"
)

func TestChan(c *testing.T) {
	ch := gochannel.NewChannel[int]()
	ctx := context.Background()
	rc, err := ch.RecvChan(ctx)
	if err != nil {
		c.Fatal(err)
	}

	sendCh, closeSend := ch.SendChan(ctx)
	defer closeSend()

	for i := 0; i < 100; i++ {
		sendCh <- i
	}

	for i := 0; i < 100; i++ {
		data := <-rc
		fmt.Println(data)
	}

	time.Sleep(time.Second * 3)
}
