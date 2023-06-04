package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"time"

	pb "github.com/centrifugal/examples/on_demand_streams/proxystreamproto"
	"google.golang.org/grpc"
)

type streamerServer struct {
	pb.UnimplementedCentrifugoProxyStreamServer
}

func (s *streamerServer) SubscribeStream(req *pb.SubscribeStreamRequest, stream pb.CentrifugoProxyStream_SubscribeStreamServer) error {
	stream.Send(&pb.Publication{})
	i := 0
	for {
		time.Sleep(time.Second)
		stream.Send(&pb.Publication{Data: []byte(`{"input": "` + strconv.Itoa(i) + `"}`)})
		i++
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	pb.RegisterCentrifugoProxyStreamServer(s, &streamerServer{})

	fmt.Println("Server listening on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
