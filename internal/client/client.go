package client

import (
	"context"

	"subpub-vk/config"
	pb "subpub-vk/internal/pb"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ClientWork(cfg *config.Config, logger *logrus.Logger) {
	port := "localhost:" + cfg.Server.Port
	conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPubSubClient(conn)

	stream, err := client.Subscribe(context.Background(), &pb.SubscribeRequest{Key: "sport"})
	if err != nil {
		logger.Fatalf("subscribe error: %v", err)
	}
	logger.Infof("subscribe success")

	_, err = client.Publish(context.Background(), &pb.PublishRequest{
		Key:  "sport",
		Data: "Ovechkin set a record for goals",
	})
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("message published")

	for {
		msg, err := stream.Recv()
		if err != nil {
			logger.Fatalf("received message error: %v", err)
		}
		logger.Infof("received: %s", msg.GetData())
	}
}
