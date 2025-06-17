package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Brand struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Country     string `json:"country"`
	Active      bool   `json:"active"`
}

var brands []Brand

func init() {
	descriptions := []string{
		"Marca premium com presença global.",
		"Referência em sustentabilidade.",
		"Foco em design minimalista e funcional.",
		"Marca líder em tecnologia de consumo.",
		"Conhecida por produtos acessíveis e duráveis.",
	}

	countries := []string{"Brasil", "Estados Unidos", "Alemanha", "Japão"}

	for i := 1; i <= 100; i++ {
		brands = append(brands, Brand{
			ID:          i,
			Name:        "Brand " + strconv.Itoa(i),
			Description: descriptions[i%len(descriptions)],
			Country:     countries[i%len(countries)],
			Active:      i%2 == 0,
		})
	}
}

func getAllBrands(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brands)
}

func getBrandByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["brandId"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	for _, b := range brands {
		if b.ID == id {
			json.NewEncoder(w).Encode(b)
			return
		}
	}

	http.Error(w, "Marca não encontrada", http.StatusNotFound)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/brands", getAllBrands).Methods("GET")
	r.HandleFunc("/brands/{brandId}", getBrandByID).Methods("GET")

	http.ListenAndServe(":8080", r)
}
