package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	pb "backend/internal/proxyproto"

	"github.com/grafana/loki/pkg/logproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	lokiPushEndpoint = "http://loki:3100/loki/api/v1/push"
	lokiGRPCAddress  = "loki:9095"
)

type streamerServer struct {
	pb.UnimplementedCentrifugoProxyServer
	lokiQuerierClient logproto.QuerierClient
}

type clientData struct {
	Query string `json:"query"`
}

func (s *streamerServer) SubscribeUnidirectional(
	req *pb.SubscribeRequest,
	stream pb.CentrifugoProxy_SubscribeUnidirectionalServer,
) error {
	var cd clientData
	err := json.Unmarshal(req.Data, &cd)
	if err != nil {
		return fmt.Errorf("error unmarshaling data: %w", err)
	}
	query := &logproto.TailRequest{
		Query: cd.Query,
	}
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	logStream, err := s.lokiQuerierClient.Tail(ctx, query)
	if err != nil {
		return fmt.Errorf("error querying Loki: %w", err)
	}

	started := time.Now()
	log.Println("unidirectional subscribe called with request", req)
	defer func() {
		log.Println("unidirectional subscribe finished, elapsed", time.Since(started))
	}()
	err = stream.Send(&pb.StreamSubscribeResponse{
		SubscribeResponse: &pb.SubscribeResponse{},
	})
	if err != nil {
		return err
	}

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		default:
			resp, err := logStream.Recv()
			if err != nil {
				return fmt.Errorf("error receiving from Loki stream: %v", err)
			}
			for _, entry := range resp.Stream.Entries {
				line := fmt.Sprintf("%s: %s", entry.Timestamp.Format("2006-01-02T15:04:05.000Z07:00"), entry.Line)
				err = stream.Send(&pb.StreamSubscribeResponse{
					Publication: &pb.Publication{Data: []byte(`{"line":"` + line + `"}`)},
				})
				if err != nil {
					return err
				}
			}
		}
	}
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type lokiPushMessage struct {
	Streams []lokiStream `json:"streams"`
}

func sendLogMessageToLoki(_ context.Context) error {
	sources := []string{"backend1", "backend2", "backend3"}
	source := sources[rand.Intn(len(sources))]
	logMessage := fmt.Sprintf("log from %s source", source)

	payload := lokiPushMessage{
		Streams: []lokiStream{
			{
				Stream: map[string]string{
					"source": source,
				},
				Values: [][]string{
					{fmt.Sprintf("%d", time.Now().UnixNano()), logMessage},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(lokiPushEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func sendLogsToLoki(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(200 * time.Millisecond):
			err := sendLogMessageToLoki(ctx)
			if err != nil {
				log.Println("error sending log to Loki:", err)
				continue
			}
		}
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	sendLogsToLoki(ctx)

	querierConn, err := grpc.Dial(lokiGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial Loki: %v", err)
	}
	querierClient := logproto.NewQuerierClient(querierConn)

	addr := ":12000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	pb.RegisterCentrifugoProxyServer(s, &streamerServer{
		lokiQuerierClient: querierClient,
	})

	log.Println("Server listening on", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
