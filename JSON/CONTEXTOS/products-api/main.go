package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type Price struct {
	Original     float64 `json:"original"`
	SpecialPrice float64 `json:"special_price"`
}

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

var products []Product

func init() {
	for i := 1; i <= 100; i++ {
		slug := "nome-do-produto-" + strconv.Itoa(i)
		products = append(products, Product{
			ID:          i,
			Name:        "Nome do produto " + strconv.Itoa(i),
			Slug:        slug,
			Description: "Descrição do Produto " + strconv.Itoa(i),
			Price: Price{
				Original:     float64(rand.Intn(100) + 1),
				SpecialPrice: float64(rand.Intn(10) + 1),
			},
			SellerID:   rand.Intn(100) + 1,
			BrandID:    rand.Intn(100) + 1,
			Categories: randomIDs(1),
			Images:     randomIDs(1),
		})
	}
}

func randomIDs(n int) []int {
	set := make(map[int]struct{})
	var result []int
	for len(result) < n {
		id := rand.Intn(100) + 1
		if _, exists := set[id]; !exists {
			set[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func getAllProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func getProductBySlug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := strings.ToLower(vars["slug"])

	for _, p := range products {
		if strings.ToLower(p.Slug) == slug {
			json.NewEncoder(w).Encode(p)
			return
		}
	}

	http.Error(w, "Produto não encontrado", http.StatusNotFound)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/products", getAllProducts).Methods("GET")
	r.HandleFunc("/products/{slug}", getProductBySlug).Methods("GET")

	http.ListenAndServe(":8080", r)
}
