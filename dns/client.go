package dns

import (
	"context"
	"fmt"

	"github.com/calvinmclean/goblin/api/gen/pb_manager"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	addr string
}

func NewGRPC(addr string) (Client, error) {
	return Client{addr}, nil
}

func (c Client) GetIP(ctx context.Context, subdomain string) (string, error) {
	conn, err := grpc.NewClient(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	client := pb_manager.NewManagerClient(conn)
	stream, err := client.GetIP(ctx, &pb_manager.GetIPRequest{
		Subdomain: subdomain,
	})
	if err != nil {
		_ = conn.Close()
		return "", fmt.Errorf("failed to GetIP: %w", err)
	}

	// keep the connection open until the context is closed
	go func() {
		select {
		case <-stream.Context().Done():
		case <-ctx.Done():
		}
		_ = conn.Close()
	}()

	resp, err := stream.Recv()
	if err != nil {
		return "", fmt.Errorf("failed to receive message: %w", err)
	}

	return resp.IpAddress, nil
}
