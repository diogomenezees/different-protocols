package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var categories = []Category{}

func init() {
	for i := 1; i <= 100; i++ {
		categories = append(categories, Category{ID: i, Name: "Category " + strconv.Itoa(i)})
	}
}

func getAllCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func getCategoryByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["categoryId"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	for _, cat := range categories {
		if cat.ID == id {
			json.NewEncoder(w).Encode(cat)
			return
		}
	}

	http.Error(w, "Categoria não encontrada", http.StatusNotFound)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/categories", getAllCategories).Methods("GET")
	r.HandleFunc("/categories/{categoryId}", getCategoryByID).Methods("GET")

	http.ListenAndServe(":8080", r)
}
