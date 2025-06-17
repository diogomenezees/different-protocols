package server

import (
	"context"
	"errors"
	"fmt"

	pb "sellers-api/proto"
)

type SellerServer struct {
	pb.UnimplementedSellerServiceServer
	sellers []*pb.Seller
}

func NewSellerServer() *SellerServer {
	var sellers []*pb.Seller
	for i := 1; i <= 100; i++ {
		sellers = append(sellers, &pb.Seller{
			Id:   int32(i),
			Name: fmt.Sprintf("Seller %d", i),
		})
	}
	return &SellerServer{sellers: sellers}
}

func (s *SellerServer) GetAllSellers(ctx context.Context, _ *pb.Empty) (*pb.SellerList, error) {
	return &pb.SellerList{Sellers: s.sellers}, nil
}

func (s *SellerServer) GetSellerByID(ctx context.Context, req *pb.SellerId) (*pb.Seller, error) {
	for _, item := range s.sellers {
		if item.Id == req.Id {
			return item, nil
		}
	}
	return nil, errors.New("seller não encontrado")
}
