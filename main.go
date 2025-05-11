package main

import (
	"context"
	"fmt"
	"subpub-vk/subpub"
)

func main() {
	sp := subpub.NewSubPub()

	hand := func(msg interface{}) {
		fmt.Println("received hand:", msg)
	}

	_, _ = sp.Subscribe("news", hand)

	sp.Publish("news", "new breaking news")

	//sub2.Unsubscribe()

	sp.Publish("sport", "football match is started")

	//sub.Unsubscribe()
	//ctx, _ := context.WithTimeout(context.Background(), time.Second)

	//defer cancel()
	sp.Close(context.Background())

}
