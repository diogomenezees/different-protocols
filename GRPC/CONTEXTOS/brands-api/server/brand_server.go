package server

import (
	"context"
	"errors"
	"fmt"

	pb "brands-api/proto"
)

type BrandServer struct {
	pb.UnimplementedBrandServiceServer
}

var brands = []*pb.Brand{}

func init() {
	descriptions := []string{
		"Marca premium com presença global.",
		"Referência em sustentabilidade.",
		"Foco em design minimalista e funcional.",
		"Marca líder em tecnologia de consumo.",
		"Conhecida por produtos acessíveis e duráveis.",
	}

	countries := []string{
		"Brasil",
		"Estados Unidos",
		"Alemanha",
		"Japão",
	}

	for i := 1; i <= 100; i++ {
		brands = append(brands, &pb.Brand{
			Id:          int32(i),
			Name:        fmt.Sprintf("Brand %d", i),
			Description: descriptions[i%len(descriptions)],
			Country:     countries[i%len(countries)],
			Active:      i%2 == 0,
		})
	}
}

func (s *BrandServer) GetAllBrands(ctx context.Context, _ *pb.Empty) (*pb.BrandList, error) {
	return &pb.BrandList{Brands: brands}, nil
}

func (s *BrandServer) GetBrandByID(ctx context.Context, req *pb.BrandRequest) (*pb.Brand, error) {
	for _, b := range brands {
		if b.Id == req.Id {
			return b, nil
		}
	}
	return nil, errors.New("marca não encontrada")
}
