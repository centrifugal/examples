package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/centrifugal/examples/unidirectional/grpc/apiproto"
	"github.com/centrifugal/examples/unidirectional/grpc/unistream"

	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:11000", "The server address in the format of host:port")
	apiAddr    = flag.String("api_addr", "localhost:10000", "The server API address")
)

func handlePush(push *unistream.Push) {
	log.Printf("push received (type %d, channel %s, data %s", push.Type, push.Channel, fmt.Sprintf("%#v", string(push.Data)))
	if push.Connect != nil {
		log.Printf("connected to a server with ID: %s", push.Connect.Client)
	} else if push.Pub != nil {
		log.Printf("new publication from channel %s: %s", push.Channel, fmt.Sprintf("%#v", string(push.Pub.Data)))
	} else if push.Join != nil {
		log.Printf("join in channel: %s (%s)", push.Channel, push.Join.Info.Client)
	} else if push.Leave != nil {
		log.Printf("Leave in channel: %s (%s)", push.Channel, push.Leave.Info.Client)
	} else {
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

func getApiClient() (apiproto.CentrifugoApiClient, func()) {
	conn, err := grpc.Dial(*apiAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	client := apiproto.NewCentrifugoApiClient(conn)
	return client, func() { conn.Close() }
}

func getChannels(client apiproto.CentrifugoApiClient) (map[string]*apiproto.ChannelInfo, error) {
	resp, err := client.Channels(context.Background(), &apiproto.ChannelsRequest{})
	if err != nil {
		return nil, fmt.Errorf("Transport level error: %v", err)
	}
	if resp.GetError() != nil {
		respError := resp.GetError()
		return nil, fmt.Errorf("Error %d (%s)", respError.Code, respError.Message)
	} else {
		return resp.Result.Channels, nil
	}
}

func askChannels(client apiproto.CentrifugoApiClient) {
	for {
		time.Sleep(time.Second)
	}
}

func publishChannels(client apiproto.CentrifugoApiClient) {
	for {
		resp, err := client.Publish(context.Background(), &apiproto.PublishRequest{
			Channel: "chat:index",
			Data:    []byte(`{"input": "test"}`),
		})
		if err != nil {
			log.Printf("Transport level error: %v", err)
		} else {
			if resp.GetError() != nil {
				respError := resp.GetError()
				log.Printf("Error %d (%s)", respError.Code, respError.Message)
			} else {
				fmt.Println("OK published")
			}
		}
		time.Sleep(time.Second)
	}
}

func main() {
	flag.Parse()

	apiClient, cancel := getApiClient()
	defer cancel()

	go publishChannels(apiClient)
	go askChannels(apiClient)

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
		channels, err := getChannels(apiClient)
		if err != nil {
			log.Fatal(err)
		}
		subs := map[string]*unistream.SubscribeRequest{}
		for ch := range channels {
			subs[ch] = &unistream.SubscribeRequest{}
		}
		subs["chat:index"] = &unistream.SubscribeRequest{}
		fmt.Println(subs)
		stream, err := client.Consume(context.Background(), &unistream.ConnectRequest{
			Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZG1pbiIsImV4cCI6MTY2NjI1OTIyMSwiaWF0IjoxNjY1NjU0NDIxfQ.a0JUKvuAXY7l0qMgeZKqWZagSYF_rP1rh8FoLNsSdvQ",
			Subs:  subs,
		})
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
