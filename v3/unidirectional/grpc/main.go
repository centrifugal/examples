package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/centrifugal/examples/unidirectional/grpc/unistream"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:11000", "The server address in the format of host:port")
)

func handlePush(push *unistream.Push) {
	log.Printf("--> push received (type %d, channel %s, data %s", push.Type, push.Channel, fmt.Sprintf("%#v", string(push.Data)))
	switch push.Type {
	case unistream.Push_CONNECT:
		var msg unistream.Connect
		err := proto.Unmarshal(push.Data, &msg)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("connected to a server with ID: %s", msg.Client)
	case unistream.Push_JOIN:
		var msg unistream.Join
		err := proto.Unmarshal(push.Data, &msg)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("client with ID: %s joined channel %s", msg.Info.Client, push.Channel)
	case unistream.Push_LEAVE:
		var msg unistream.Leave
		err := proto.Unmarshal(push.Data, &msg)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("client with ID: %s left channel %s", msg.Info.Client, push.Channel)
	case unistream.Push_PUBLICATION:
		var p unistream.Publication
		err := proto.Unmarshal(push.Data, &p)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("new publication from channel %s: %s", push.Channel, fmt.Sprintf("%#v", string(p.Data)))
	case unistream.Push_DISCONNECT:
		var msg unistream.Disconnect
		err := proto.Unmarshal(push.Data, &msg)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("disconnected from a server: %s", msg.Reason)
	default:
		log.Println("push type handling not implemented")
	}
}

func handleStream(stream unistream.CentrifugoUniStream_ConsumeClient) error {
	for {
		push, err := stream.Recv()
		if err != nil {
			return err
		}
		handlePush(push)
	}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	client := unistream.NewCentrifugoUniStreamClient(conn)

	numFailureAttempts := 0
	for {
		time.Sleep(time.Duration(numFailureAttempts) * time.Second)
		log.Println("establishing a unidirectional stream")
		stream, err := client.Consume(context.Background(), &unistream.ConnectRequest{})
		if err != nil {
			log.Printf("error establishing stream: %v", err)
			numFailureAttempts++
			continue
		}
		log.Println("stream established")
		numFailureAttempts = 0
		err = handleStream(stream)
		if err != nil {
			log.Printf("error handling stream: %v", err)
			time.Sleep(time.Second)
		}
	}
}
