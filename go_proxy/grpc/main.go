package main

import (
	"log"
	"net"

	"github.com/centrifugal/examples/proxy/grpc/proxyproto"
	"google.golang.org/grpc"
)

func main() {
	const port = ":10001"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("cannot listen port - ", err)
	}
	s := grpc.NewServer()
	proxyproto.RegisterCentrifugoProxyServer(s, &proxyproto.Server{})

	log.Printf("Starting gRPC server on port %s", port)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve grpc server: %v", err)
	}
}
