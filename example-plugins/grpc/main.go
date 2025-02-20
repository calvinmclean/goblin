package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"grpc-example/gen/pb_example"

	"google.golang.org/grpc"
)

//go:generate protoc -I. --go_out=./gen --go-grpc_out=./gen example.proto

type Server struct {
	pb_example.UnimplementedExampleServer
}

func (s Server) Run(ctx context.Context, grpcAddr string) error {
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to create Listener: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb_example.RegisterExampleServer(grpcServer, s)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
		wg.Done()
	}()

	go func() {
		err = grpcServer.Serve(listener)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return nil
}

func (s Server) Hello(context.Context, *pb_example.HelloRequest) (*pb_example.HelloResponse, error) {
	return &pb_example.HelloResponse{
		Greeting: "Hello, World!",
	}, nil
}
