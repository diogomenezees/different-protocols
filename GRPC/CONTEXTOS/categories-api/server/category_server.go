package server

import (
	"context"
	"errors"
	"fmt"

	pb "categories-api/proto"
)

type CategoryServer struct {
	pb.UnimplementedCategoryServiceServer
	categories []*pb.Category
}

func NewCategoryServer() *CategoryServer {
	var categories []*pb.Category
	for i := 1; i <= 100; i++ {
		categories = append(categories, &pb.Category{
			Id:   int32(i),
			Name: fmt.Sprintf("Category %d", i),
		})
	}
	return &CategoryServer{categories: categories}
}

func (s *CategoryServer) GetAllCategories(ctx context.Context, _ *pb.Empty) (*pb.CategoryList, error) {
	return &pb.CategoryList{Categories: s.categories}, nil
}

func (s *CategoryServer) GetCategoryByID(ctx context.Context, req *pb.CategoryId) (*pb.Category, error) {
	for _, c := range s.categories {
		if c.Id == req.Id {
			return c, nil
		}
	}
	return nil, errors.New("categoria não encontrada")
}
