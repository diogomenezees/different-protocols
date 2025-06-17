package main

import (
	"log"
	"net"

	pb "brands-api/proto"
	"brands-api/server"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterBrandServiceServer(s, &server.BrandServer{})

	log.Println("Brand gRPC server running on port 8080")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
