package main

import (
	"log"
	"net"

	pb "categories-api/proto"
	"categories-api/server"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Erro ao escutar: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCategoryServiceServer(s, server.NewCategoryServer())

	log.Println("Servidor gRPC de categorias rodando na porta 8080")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Falha ao servir: %v", err)
	}
}
