package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	brandpb "bff/proto/brand"
	categorypb "bff/proto/category"
	imagepb "bff/proto/image"
	productpb "bff/proto/product"
	sellerpb "bff/proto/seller"
)

const (
	brandAPI    = "brands-grpc-api:8080"
	categoryAPI = "categories-grpc-api:8080"
	imageAPI    = "images-grpc-api:8080"
	productAPI  = "products-grpc-api:8080"
	sellerAPI   = "sellers-grpc-api:8080"
)

type ProductResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Price       any    `json:"price"`
	Seller      any    `json:"seller"`
	Brand       any    `json:"brand"`
	Categories  []any  `json:"categories"`
	Images      []any  `json:"images"`
}

var (
	productConn  *grpc.ClientConn
	brandConn    *grpc.ClientConn
	sellerConn   *grpc.ClientConn
	categoryConn *grpc.ClientConn
	imageConn    *grpc.ClientConn

	productClient  productpb.ProductServiceClient
	brandClient    brandpb.BrandServiceClient
	sellerClient   sellerpb.SellerServiceClient
	categoryClient categorypb.CategoryServiceClient
	imageClient    imagepb.ImageServiceClient
)

func initClients() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	productConn, err = grpc.DialContext(ctx, productAPI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	brandConn, err = grpc.DialContext(ctx, brandAPI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	sellerConn, err = grpc.DialContext(ctx, sellerAPI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	categoryConn, err = grpc.DialContext(ctx, categoryAPI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	imageConn, err = grpc.DialContext(ctx, imageAPI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	productClient = productpb.NewProductServiceClient(productConn)
	brandClient = brandpb.NewBrandServiceClient(brandConn)
	sellerClient = sellerpb.NewSellerServiceClient(sellerConn)
	categoryClient = categorypb.NewCategoryServiceClient(categoryConn)
	imageClient = imagepb.NewImageServiceClient(imageConn)

	return nil
}

func closeClients() {
	productConn.Close()
	brandConn.Close()
	sellerConn.Close()
	categoryConn.Close()
	imageConn.Close()
}

func fetchProduct(slug string, client productpb.ProductServiceClient) (*productpb.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return client.GetProductBySlug(ctx, &productpb.Slug{Slug: slug})
}

func fetchBrand(id int32, client brandpb.BrandServiceClient) (*brandpb.Brand, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return client.GetBrandByID(ctx, &brandpb.BrandRequest{Id: id})
}

func fetchSeller(id int32, client sellerpb.SellerServiceClient) (*sellerpb.Seller, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return client.GetSellerByID(ctx, &sellerpb.SellerId{Id: id})
}

func fetchCategory(id int32, client categorypb.CategoryServiceClient) (*categorypb.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return client.GetCategoryByID(ctx, &categorypb.CategoryId{Id: id})
}

func fetchImage(id int32, client imagepb.ImageServiceClient) (*imagepb.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return client.GetImageByID(ctx, &imagepb.ImageId{Id: id})
}

func GetProductSequential(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := strings.ToLower(vars["slug"])

	prod, err := fetchProduct(slug, productClient)
	if err != nil {
		http.Error(w, "Produto não encontrado", http.StatusNotFound)
		return
	}

	brand, _ := fetchBrand(prod.BrandId, brandClient)

	seller, _ := fetchSeller(prod.SellerId, sellerClient)

	categories := []any{}
	for _, cid := range prod.Categories {
		if cat, err := fetchCategory(cid, categoryClient); err == nil {
			categories = append(categories, cat)
		}
	}

	images := []any{}
	for _, iid := range prod.Images {
		if img, err := fetchImage(iid, imageClient); err == nil {
			images = append(images, img)
		}
	}

	resp := ProductResponse{
		ID:          int(prod.Id),
		Name:        prod.Name,
		Slug:        prod.Slug,
		Description: prod.Description,
		Price:       prod.Price,
		Seller:      seller,
		Brand:       brand,
		Categories:  categories,
		Images:      images,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GetProductParallel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := strings.ToLower(vars["slug"])

	prod, err := fetchProduct(slug, productClient)
	if err != nil {
		http.Error(w, "Produto não encontrado", http.StatusNotFound)
		return
	}

	var wg sync.WaitGroup
	var brand any
	var seller any
	categories := make([]any, len(prod.Categories))
	images := make([]any, len(prod.Images))

	wg.Add(1)
	go func() {
		defer wg.Done()
		brand, _ = fetchBrand(prod.BrandId, brandClient)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		seller, _ = fetchSeller(prod.SellerId, sellerClient)
	}()

	for i, cid := range prod.Categories {
		wg.Add(1)
		go func(i int, cid int32) {
			defer wg.Done()
			if cat, err := fetchCategory(cid, categoryClient); err == nil {
				categories[i] = cat
			}
		}(i, cid)
	}

	for i, iid := range prod.Images {
		wg.Add(1)
		go func(i int, iid int32) {
			defer wg.Done()
			if img, err := fetchImage(iid, imageClient); err == nil {
				images[i] = img
			}
		}(i, iid)
	}

	wg.Wait()

	resp := ProductResponse{
		ID:          int(prod.Id),
		Name:        prod.Name,
		Slug:        prod.Slug,
		Description: prod.Description,
		Price:       prod.Price,
		Seller:      seller,
		Brand:       brand,
		Categories:  categories,
		Images:      images,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	if err := initClients(); err != nil {
		log.Fatalf("Erro ao inicializar clientes gRPC: %v", err)
	}
	defer closeClients()

	r := mux.NewRouter()
	r.HandleFunc("/sequencial/{slug}", GetProductSequential).Methods("GET")
	r.HandleFunc("/paralelo/{slug}", GetProductParallel).Methods("GET")

	log.Println("Servidor BFF rodando na porta 8080")
	http.ListenAndServe(":8080", r)
}
