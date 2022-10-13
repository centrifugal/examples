package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/centrifugal/examples/unidirectional/grpc/apiproto"
	"github.com/centrifugal/examples/unidirectional/grpc/unistream"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:11000", "The server address in the format of host:port")
	apiAddr    = flag.String("api_addr", "localhost:10000", "The server API address")
)

func handlePush(push *unistream.Push) {
	if push.Connect != nil {
		log.Printf("connected to a server with ID: %s", push.Connect.Client)
	} else if push.Pub != nil {
		log.Printf("new publication from channel %s: %s", push.Channel, fmt.Sprintf("%#v", string(push.Pub.Data)))
	} else if push.Join != nil {
		log.Printf("join in channel: %s (%s)", push.Channel, push.Join.Info.Client)
	} else if push.Leave != nil {
		log.Printf("Leave in channel: %s (%s)", push.Channel, push.Leave.Info.Client)
	} else {
		log.Printf("unimplemented push: %#v", push)
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

func publishChannel(client apiproto.CentrifugoApiClient) {
	for {
		data, _ := json.Marshal(map[string]string{
			"input": fmt.Sprintf("test_%d", time.Now().Unix()),
		})
		resp, err := client.Publish(context.Background(), &apiproto.PublishRequest{
			Channel: "test_channel",
			Data:    data,
		})
		if err != nil {
			log.Printf("Transport level error: %v", err)
		} else {
			if resp.GetError() != nil {
				respError := resp.GetError()
				log.Printf("Error %d (%s)", respError.Code, respError.Message)
			} else {
				fmt.Println("Publish OK")
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func main() {
	flag.Parse()

	// NOTE, that you should never reveal token secret key to your users!
	// It should only be known by your app backend and Centrifugo.
	// In real app you may get the connection token from the outside of the program.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "example_user",
		"channels": []string{"test_channel"},
	})
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		log.Fatal(err)
	}

	apiClient, cancel := getApiClient()
	defer cancel()

	go publishChannel(apiClient)

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
		stream, err := client.Consume(context.Background(), &unistream.ConnectRequest{
			Token: tokenString,
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
