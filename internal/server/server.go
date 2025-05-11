package server

import (
	"context"
	pb "subpub-vk/internal/pb"
	"subpub-vk/subpub"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PubSubServer struct {
	pb.UnimplementedPubSubServer
	sp     subpub.SubPub
	logger *logrus.Logger
	mu     sync.RWMutex
}

func NewServer(sp subpub.SubPub, logger *logrus.Logger) *PubSubServer {
	return &PubSubServer{
		sp:     sp,
		logger: logger,
	}
}

func (s *PubSubServer) Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) error {
	key := req.GetKey()

	sub, err := s.sp.Subscribe(key, func(msg interface{}) {

		if err := stream.Send(&pb.Event{Data: msg.(string)}); err != nil {
			s.logger.Errorf("send to subject error (key = %s): %v", key, err)
		}
	})

	if err != nil {
		s.logger.Errorf("subscription error (key = %s): %v", key, err)
		return status.Error(codes.Internal, "failed to subscribe")
	}

	<-stream.Context().Done()
	sub.Unsubscribe()
	return nil
}

func (s *PubSubServer) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	key := req.GetKey()
	data := req.GetData()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Infof("Publishing to subject %s: %s", key, data)

	err := s.sp.Publish(key, data)
	if err != nil {
		s.logger.Errorf("error publishing  message for key %v: %v", key, err)
		return nil, status.Errorf(codes.Internal, "publish error %v", err)
	}
	return &emptypb.Empty{}, nil
}
