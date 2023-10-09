package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"time"

	pb "github.com/centrifugal/examples/on_demand_streams/proxyproto"
	"google.golang.org/grpc"
)

type streamerServer struct {
	pb.UnimplementedCentrifugoProxyServer
}

func (s *streamerServer) SubscribeUnidirectional(
	req *pb.SubscribeRequest,
	stream pb.CentrifugoProxy_SubscribeUnidirectionalServer,
) error {
	started := time.Now()
	fmt.Println("unidirectional subscribe called with request", req)
	defer func() {
		fmt.Println("unidirectional subscribe finished, elapsed", time.Since(started))
	}()
	stream.Send(&pb.StreamSubscribeResponse{
		SubscribeResponse: &pb.SubscribeResponse{},
	})
	i := 0
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-time.After(1000 * time.Millisecond):
		}
		pub := &pb.Publication{Data: []byte(`{"input": "` + strconv.Itoa(i) + `"}`)}
		stream.Send(&pb.StreamSubscribeResponse{Publication: pub})
		i++
		if i >= 20 {
			break
		}
	}
	return nil
}

type clientData struct {
	Input string `json:"input"`
}

func (s *streamerServer) SubscribeBidirectional(
	stream pb.CentrifugoProxy_SubscribeBidirectionalServer,
) error {
	started := time.Now()
	fmt.Println("bidirectional subscribe called")
	defer func() {
		fmt.Println("bidirectional subscribe finished, elapsed", time.Since(started))
	}()
	// First message always contains SubscribeRequest.
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	fmt.Println("subscribe request received", req.SubscribeRequest)
	stream.Send(&pb.StreamSubscribeResponse{
		SubscribeResponse: &pb.SubscribeResponse{},
	})
	// The following messages contain publications from client.
	for {
		req, err := stream.Recv()
		if err != nil {
			fmt.Println(err)
			return err
		}
		data := req.Publication.Data
		fmt.Println("data from client", string(data))
		var cd clientData
		err = json.Unmarshal(data, &cd)
		if err != nil {
			return nil
		}
		pub := &pb.Publication{Data: []byte(`{"input": "` + cd.Input + `"}`)}
		stream.Send(&pb.StreamSubscribeResponse{Publication: pub})
	}
}

func main() {
	addr := ":12000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	pb.RegisterCentrifugoProxyServer(s, &streamerServer{})

	fmt.Println("Server listening on", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
