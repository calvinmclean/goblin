package server

import (
	"context"
	"dns-plugin-thing/api/gen/pb_manager"
	"dns-plugin-thing/dns"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type Server struct {
	pb_manager.UnimplementedManagerServer
	mgr dns.Manager
}

func New(mgr dns.Manager) Server {
	return Server{
		pb_manager.UnimplementedManagerServer{},
		mgr,
	}
}

func (s Server) Run(ctx context.Context, grpcAddr string) error {
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to create Listener: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb_manager.RegisterManagerServer(grpcServer, s)

	var wg sync.WaitGroup
	wg.Add(3)

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

	go func() {
		err = s.mgr.RunDNS(ctx)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return nil
}

// GetIP: Allocates an IP and releases it when the client disconnects
func (s Server) GetIP(req *pb_manager.GetIPRequest, stream pb_manager.Manager_GetIPServer) error {
	ip, err := s.mgr.GetIP(stream.Context(), req.Subdomain)
	if err != nil {
		return err
	}

	err = stream.Send(&pb_manager.GetIPResponse{IpAddress: ip})
	if err != nil {
		return err
	}

	// keep context open until the client closes it
	<-stream.Context().Done()

	return nil
}
