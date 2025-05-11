package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"subpub-vk/config"
	"subpub-vk/internal/client"
	pb "subpub-vk/internal/pb"
	"subpub-vk/internal/server"
	log "subpub-vk/logger"
	"subpub-vk/subpub"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.InitConfig()
	logger := log.InitLogger(cfg)

	logger.Infof("Config is loaded: %v", cfg)

	sp := subpub.NewSubPub()

	grpcServer := grpc.NewServer()
	pb.RegisterPubSubServer(grpcServer, server.NewServer(sp, logger))

	lis, err := net.Listen("tcp", "localhost:"+cfg.Server.Port)
	if err != nil {
		logger.Fatalf("listen server error: %v", err)
	}

	go func() {
		logger.Infof("Starting gRPC server on port %s", cfg.Server.Port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("serve error: %v", err)
		}
	}()

	go client.ClientWork(cfg, logger)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop() // остановка сервера

	if err := sp.Close(ctx); err != nil {
		logger.Errorf("Error closing subpub: %v", err)
	}

	logger.Info("Server stopped gracefully")

}
