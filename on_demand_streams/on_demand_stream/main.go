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

func (s *streamerServer) Consume(req *pb.SubscribeRequest, stream pb.CentrifugoProxyStream_ConsumeServer) error {
	stream.Send(&pb.Response{})
	i := 0
	for {
		time.Sleep(time.Second)
		pub := &pb.Publication{Data: []byte(`{"input": "` + strconv.Itoa(i) + `"}`)}
		stream.Send(&pb.Response{Publication: pub})
		i++
	}
}

func main() {
	addr := ":12000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	pb.RegisterCentrifugoProxyStreamServer(s, &streamerServer{})

	fmt.Println("Server listening on", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
