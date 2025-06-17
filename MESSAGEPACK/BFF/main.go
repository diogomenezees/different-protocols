package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	brandAPI    = "http://brands-msgpack-api:8080/brands/%d"
	categoryAPI = "http://categories-msgpack-api:8080/categories/%d"
	imageAPI    = "http://images-msgpack-api:8080/images/%d"
	productAPI  = "http://products-msgpack-api:8080/products/%s"
	sellerAPI   = "http://sellers-msgpack-api:8080/sellers/%d"
)

type Price struct {
	Original     float64 `json:"original" msgpack:"original"`
	SpecialPrice float64 `json:"special_price" msgpack:"special_price"`
}

// Response para MessagePack
type Product struct {
	ID          int    `msgpack:"id"`
	Name        string `msgpack:"name"`
	Slug        string `msgpack:"slug"`
	Description string `msgpack:"description"`
	Price       Price  `msgpack:"price"`
	SellerID    int    `msgpack:"seller_id"`
	BrandID     int    `msgpack:"brand_id"`
	Categories  []int  `msgpack:"categories"`
	Images      []int  `msgpack:"images"`
}

// Response para resposta
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

var clientMsgPack = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	},
	Timeout: 10 * time.Second, // timeout total da requisição
}

func fetchMsgPack[T any](url string, target *T) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/x-msgpack")

	resp, err := clientMsgPack.Do(req) // usa o client global
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return msgpack.Unmarshal(body, target)
}

func EnrichProductSequential(slug string) (*ProductResponse, error) {
	var product Product
	if err := fetchMsgPack(fmt.Sprintf(productAPI, slug), &product); err != nil {
		return nil, err
	}

	var seller map[string]interface{}
	var brand map[string]interface{}
	var categories []interface{}
	var images []interface{}

	fetchMsgPack(fmt.Sprintf(sellerAPI, product.SellerID), &seller)
	fetchMsgPack(fmt.Sprintf(brandAPI, product.BrandID), &brand)

	for _, id := range product.Categories {
		var c map[string]interface{}
		fetchMsgPack(fmt.Sprintf(categoryAPI, id), &c)
		categories = append(categories, c)
	}

	for _, id := range product.Images {
		var img map[string]interface{}
		fetchMsgPack(fmt.Sprintf(imageAPI, id), &img)
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
	if err := fetchMsgPack(fmt.Sprintf(productAPI, slug), &product); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var seller, brand map[string]interface{}
	var categories = make([]interface{}, len(product.Categories))
	var images = make([]interface{}, len(product.Images))

	wg.Add(2 + len(product.Categories) + len(product.Images))

	go func() {
		defer wg.Done()
		fetchMsgPack(fmt.Sprintf(sellerAPI, product.SellerID), &seller)
	}()

	go func() {
		defer wg.Done()
		fetchMsgPack(fmt.Sprintf(brandAPI, product.BrandID), &brand)
	}()

	for i, id := range product.Categories {
		go func(i int, id int) {
			defer wg.Done()
			var c map[string]interface{}
			fetchMsgPack(fmt.Sprintf(categoryAPI, id), &c)
			mu.Lock()
			categories[i] = c
			mu.Unlock()
		}(i, id)
	}

	for i, id := range product.Images {
		go func(i int, id int) {
			defer wg.Done()
			var img map[string]interface{}
			fetchMsgPack(fmt.Sprintf(imageAPI, id), &img)
			mu.Lock()
			images[i] = img
			mu.Unlock()
		}(i, id)
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
		vars := mux.Vars(r)
		result, err := EnrichProductSequential(vars["slug"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	r.HandleFunc("/paralelo/{slug}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		result, err := EnrichProductParallel(vars["slug"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.ListenAndServe(":8080", r)
}
