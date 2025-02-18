package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"dns-plugin-thing/api/gen/pb_manager"
	"dns-plugin-thing/dns"
	"dns-plugin-thing/plugins"
	"dns-plugin-thing/server"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	app := &cli.App{
		Name: "dns-plugin-thing",
		Commands: []*cli.Command{
			{
				Name:        "client",
				Description: "run client",
				Action:      runClient,
			},
			{
				Name:        "serve",
				Description: "run server",
				Action:      runServer,
			},
			{
				Name:        "example",
				Description: "run example",
				Action:      runExample,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runExample(c *cli.Context) error {
	dnsMgr := dns.New(".gotest.")

	go func() {
		err := runPlugin(dnsMgr, "./plugins/examples/hello-world.so", "helloworld", 0)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)

		err := runPlugin(dnsMgr, "./plugins/examples/howdy-world.so", "howdy", 0)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		time.Sleep(15 * time.Second)

		err := runPlugin(dnsMgr, "./plugins/examples/howdy-world.so", "howdynew", 0)
		if err != nil {
			panic(err)
		}
	}()

	ctx := context.Background()

	server := server.New(dnsMgr)
	go func() {
		err := server.Run(ctx, ":50051")
		if err != nil {
			log.Fatalf("error running GRPC server: %v", err)
		}
	}()

	err := dnsMgr.RunDNS(ctx, "127.0.0.1:5154")
	if err != nil {
		return fmt.Errorf("error running DNS server: %w", err)
	}
	return nil
}

func runServer(c *cli.Context) error {
	dnsMgr := dns.New(".gotest.")

	ctx := context.Background()

	server := server.New(dnsMgr)
	go func() {
		err := server.Run(ctx, ":50051")
		if err != nil {
			log.Fatalf("error running GRPC server: %v", err)
		}
	}()

	err := dnsMgr.RunDNS(ctx, "127.0.0.1:5154")
	if err != nil {
		return fmt.Errorf("error running DNS server: %w", err)
	}
	return nil
}

func runPlugin(dnsMgr dns.Manager, fname, hostname string, timeout time.Duration) error {
	ctx := context.Background()
	if timeout != 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ctx = timeoutCtx
	}

	log.Printf("starting plugin: %q", hostname)
	run, err := plugins.Load(fname)
	if err != nil {
		return fmt.Errorf("error loading plugin: %w", err)
	}

	ip, err := dnsMgr.GetIP(ctx, hostname)
	if err != nil {
		return fmt.Errorf("error getting IP: %w", err)
	}

	err = run(ctx, ip)
	log.Printf("stopped plugin: %q", hostname)
	return err
}

func runClient(c *cli.Context) error {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	subdomain := c.Args().First()
	if subdomain == "" {
		subdomain = "test"
	}

	client := pb_manager.NewManagerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
