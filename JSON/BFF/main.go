package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	brandAPI    = "http://brands-api:8080/brands/%d"
	categoryAPI = "http://categories-api:8080/categories/%d"
	imageAPI    = "http://images-api:8080/images/%d"
	productAPI  = "http://products-api:8080/products/%s"
	sellerAPI   = "http://sellers-api:8080/sellers/%d"
)

type Price struct {
	Original     float64 `json:"original"`
	SpecialPrice float64 `json:"special_price"`
}

// Response original
type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Price       Price  `json:"price"`
	SellerID    int    `json:"seller_id"`
	BrandID     int    `json:"brand_id"`
	Categories  []int  `json:"categories"`
	Images      []int  `json:"images"`
}

// Response de resposta
type ProductResponse struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description"`
	Price       Price                  `json:"price"`
	Seller      map[string]interface{} `json:"seller"`
	Brand       map[string]interface{} `json:"brand"`
	Categories  []interface{}          `json:"categories"`
	Images      []interface{}          `json:"images"`
}

var client = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	},
	Timeout: 10 * time.Second, // Timeout total para a requisição
}

func fetch[T any](url string, target *T) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func EnrichProductSequential(slug string) (*ProductResponse, error) {
	var product Product
	if err := fetch(fmt.Sprintf(productAPI, slug), &product); err != nil {
		return nil, err
	}

	var seller map[string]interface{}
	var brand map[string]interface{}
	var categories []interface{}
	var images []interface{}

	fetch(fmt.Sprintf(sellerAPI, product.SellerID), &seller)
	fetch(fmt.Sprintf(brandAPI, product.BrandID), &brand)

	for _, id := range product.Categories {
		var c map[string]interface{}
		fetch(fmt.Sprintf(categoryAPI, id), &c)
		categories = append(categories, c)
	}

	for _, id := range product.Images {
		var img map[string]interface{}
		fetch(fmt.Sprintf(imageAPI, id), &img)
		images = append(images, img)
	}

	var response ProductResponse

	response.ID = product.ID
	response.Name = product.Name
	response.Slug = product.Slug
	response.Description = product.Description
	response.Price = product.Price
	response.Seller = seller
	response.Brand = brand
	response.Categories = categories
	response.Images = images

	return &response, nil
}

func EnrichProductParallel(slug string) (*ProductResponse, error) {
	var product Product
	if err := fetch(fmt.Sprintf(productAPI, slug), &product); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	var seller, brand map[string]interface{}
	var categories, images []interface{}

	// Abertura de 8 goroutines
	wg.Add(2 + len(product.Categories) + len(product.Images))

	// Para o Seller
	go func() {
		defer wg.Done()
		var s map[string]interface{}
		fetch(fmt.Sprintf(sellerAPI, product.SellerID), &s)
		mu.Lock()
		seller = s
		mu.Unlock()
	}()

	// Retornar a Marca
	go func() {
		defer wg.Done()
		var b map[string]interface{}
		fetch(fmt.Sprintf(brandAPI, product.BrandID), &b)
		mu.Lock()
		brand = b
		mu.Unlock()
	}()

	// Para cada categoria
	for _, id := range product.Categories {
		go func(id int) {
			defer wg.Done()
			var c map[string]interface{}
			fetch(fmt.Sprintf(categoryAPI, id), &c)
			mu.Lock()
			categories = append(categories, c)
			mu.Unlock()
		}(id)
	}

	// Para cada imagem
	for _, id := range product.Images {
		go func(id int) {
			defer wg.Done()
			var img map[string]interface{}
			fetch(fmt.Sprintf(imageAPI, id), &img)
			mu.Lock()
			images = append(images, img)
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	var response ProductResponse

	response.ID = product.ID
	response.Name = product.Name
	response.Slug = product.Slug
	response.Description = product.Description
	response.Price = product.Price
	response.Seller = seller
	response.Brand = brand
	response.Categories = categories
	response.Images = images

	return &response, nil
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/sequencial/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := mux.Vars(r)["slug"]
		product, err := EnrichProductSequential(slug)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	})

	r.HandleFunc("/paralelo/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := mux.Vars(r)["slug"]
		product, err := EnrichProductParallel(slug)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	})

	http.ListenAndServe(":8080", r)
}
