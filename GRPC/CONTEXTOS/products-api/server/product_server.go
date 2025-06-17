package server

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	pb "products-api/proto"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
	products []*pb.Product
}

func NewProductServer() *ProductServer {
	var products []*pb.Product
	for i := 1; i <= 100; i++ {
		products = append(products, &pb.Product{
			Id:          int32(i),
			Name:        fmt.Sprintf("Product %d", i),
			Slug:        fmt.Sprintf("nome-do-produto-%d", i),
			Description: fmt.Sprintf("Descrição do produto %d", i),
			Price: &pb.Price{
				Original:     float32(rand.Intn(100) + 1),
				SpecialPrice: float32(rand.Intn(100) + 1),
			},
			SellerId:   int32(rand.Intn(100) + 1),
			BrandId:    int32(rand.Intn(100) + 1),
			Categories: randomIDs(1),
			Images:     randomIDs(1),
		})
	}
	return &ProductServer{products: products}
}

func randomIDs(n int) []int32 {
	set := make(map[int]struct{})
	var result []int32

	for len(result) < n {
		id := rand.Intn(100) + 1
		if _, exists := set[id]; !exists {
			set[id] = struct{}{}
			result = append(result, int32(id))
		}
	}
	return result
}

func (s *ProductServer) GetAllProducts(ctx context.Context, _ *pb.Empty) (*pb.ProductList, error) {
	return &pb.ProductList{Products: s.products}, nil
}

func (s *ProductServer) GetProductBySlug(ctx context.Context, req *pb.Slug) (*pb.Product, error) {
	slug := strings.ToLower(req.Slug)
	for _, item := range s.products {
		if strings.ToLower(item.Slug) == slug {
			return item, nil
		}
	}
	return nil, errors.New("product não encontrado")
}
