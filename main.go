package main

import (
	"context"
	"fmt"
	"time"

	"subpub-vk/subpub"
)

func main() {
	sp := subpub.NewSubPub()

	sub, _ := sp.Subscribe("news", func(msg interface{}) {
		fmt.Println("received sub1:", msg)
	})

	sp.Publish("news", "hello world")

	sub.Unsubscribe()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sp.Close(ctx)

}
