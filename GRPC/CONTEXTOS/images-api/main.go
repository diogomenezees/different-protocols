package main

import (
	"log"
	"net"

	pb "images-api/proto"
	"images-api/server"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Erro ao escutar: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterImageServiceServer(s, server.NewImageServer())

	log.Println("Servidor gRPC de image rodando na porta 8080")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Falha ao servir: %v", err)
	}
}
