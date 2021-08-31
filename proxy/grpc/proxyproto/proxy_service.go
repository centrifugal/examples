package proxyproto

import (
	context "context"
	"log"
)

type Server struct{}

func (s *Server) Connect(ctx context.Context, request *ConnectRequest) (*ConnectResponse, error) {
	log.Println("Connect..")
	return &ConnectResponse{}, nil
}

func (s *Server) Refresh(ctx context.Context, request *RefreshRequest) (*RefreshResponse, error) {
	log.Println("Refresh..")
	return &RefreshResponse{}, nil
}

func (s *Server) Subscribe(ctx context.Context, request *SubscribeRequest) (*SubscribeResponse, error) {
	log.Println("Subscribe..")
	return &SubscribeResponse{}, nil
}

func (s *Server) Publish(ctx context.Context, request *PublishRequest) (*PublishResponse, error) {
	log.Println("Publish..")
	log.Println(string(request.Data))
	return &PublishResponse{}, nil
}

func (s *Server) RPC(ctx context.Context, request *RPCRequest) (*RPCResponse, error) {
	log.Println("RPC..")
	return &RPCResponse{}, nil
}

func (s *Server) mustEmbedUnimplementedCentrifugoProxyServer() {
	panic("implement me")
}
