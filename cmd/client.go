package cmd

import (
	"context"
	"dns-plugin-thing/api/gen/pb_manager"
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ClientCmd = &cli.Command{
	Name:        "client",
	Description: "run client",
	Action:      runClient,
}

func runClient(c *cli.Context) error {
	conn, err := grpc.NewClient(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	subdomain := c.Args().First()
	if subdomain == "" {
		subdomain = "test"
	}

	client := pb_manager.NewManagerClient(conn)
	ctx, cancel := context.WithTimeout(c.Context, 5*time.Second)
	defer cancel()

	stream, err := client.GetIP(ctx, &pb_manager.GetIPRequest{
		Subdomain: subdomain,
	})
	if err != nil {
		return fmt.Errorf("failed to GetIP: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive message: %w", err)
	}
	log.Printf("Got IP: %s", resp.IpAddress)

	<-stream.Context().Done()

	return nil
}
