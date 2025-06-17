package server

import (
	"context"
	"errors"
	"fmt"

	pb "images-api/proto"
)

type ImageServer struct {
	pb.UnimplementedImageServiceServer
	images []*pb.Image
}

func NewImageServer() *ImageServer {
	var images []*pb.Image
	for i := 1; i <= 100; i++ {
		images = append(images, &pb.Image{
			Id:  int32(i),
			Url: fmt.Sprintf("https://example.com/image%d.jpg", i),
		})
	}
	return &ImageServer{images: images}
}

func (s *ImageServer) GetAllImages(ctx context.Context, _ *pb.Empty) (*pb.ImageList, error) {
	return &pb.ImageList{Images: s.images}, nil
}

func (s *ImageServer) GetImageByID(ctx context.Context, req *pb.ImageId) (*pb.Image, error) {
	for _, item := range s.images {
		if item.Id == req.Id {
			return item, nil
		}
	}
	return nil, errors.New("image não encontrado")
}
