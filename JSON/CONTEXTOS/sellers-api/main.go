package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Seller struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var sellers = []Seller{
	{ID: 1, Name: "Seller A"},
	{ID: 2, Name: "Seller B"},
	{ID: 3, Name: "Seller C"},
	{ID: 4, Name: "Seller D"},
	{ID: 5, Name: "Seller E"},
	{ID: 6, Name: "Seller F"},
	{ID: 7, Name: "Seller G"},
	{ID: 8, Name: "Seller H"},
	{ID: 9, Name: "Seller I"},
	{ID: 10, Name: "Seller J"},
	{ID: 11, Name: "Seller K"},
	{ID: 12, Name: "Seller L"},
	{ID: 13, Name: "Seller M"},
	{ID: 14, Name: "Seller N"},
	{ID: 15, Name: "Seller O"},
	{ID: 16, Name: "Seller P"},
	{ID: 17, Name: "Seller Q"},
	{ID: 18, Name: "Seller R"},
	{ID: 19, Name: "Seller S"},
	{ID: 20, Name: "Seller T"},
	{ID: 21, Name: "Seller U"},
	{ID: 22, Name: "Seller V"},
	{ID: 23, Name: "Seller W"},
	{ID: 24, Name: "Seller X"},
	{ID: 25, Name: "Seller Y"},
	{ID: 26, Name: "Seller Z"},
	{ID: 27, Name: "Seller Alpha"},
	{ID: 28, Name: "Seller Beta"},
	{ID: 29, Name: "Seller Gamma"},
	{ID: 30, Name: "Seller Delta"},
	{ID: 31, Name: "Seller Epsilon"},
	{ID: 32, Name: "Seller Zeta"},
	{ID: 33, Name: "Seller Eta"},
	{ID: 34, Name: "Seller Theta"},
	{ID: 35, Name: "Seller Iota"},
	{ID: 36, Name: "Seller Kappa"},
	{ID: 37, Name: "Seller Lambda"},
	{ID: 38, Name: "Seller Mu"},
	{ID: 39, Name: "Seller Nu"},
	{ID: 40, Name: "Seller Xi"},
	{ID: 41, Name: "Seller Omicron"},
	{ID: 42, Name: "Seller Pi"},
	{ID: 43, Name: "Seller Rho"},
	{ID: 44, Name: "Seller Sigma"},
	{ID: 45, Name: "Seller Tau"},
	{ID: 46, Name: "Seller Upsilon"},
	{ID: 47, Name: "Seller Phi"},
	{ID: 48, Name: "Seller Chi"},
	{ID: 49, Name: "Seller Psi"},
	{ID: 50, Name: "Seller Omega"},
	{ID: 51, Name: "Seller Nova"},
	{ID: 52, Name: "Seller Orbit"},
	{ID: 53, Name: "Seller Stellar"},
	{ID: 54, Name: "Seller Nebula"},
	{ID: 55, Name: "Seller Quasar"},
	{ID: 56, Name: "Seller Vortex"},
	{ID: 57, Name: "Seller Eclipse"},
	{ID: 58, Name: "Seller Blaze"},
	{ID: 59, Name: "Seller Ember"},
	{ID: 60, Name: "Seller Frost"},
	{ID: 61, Name: "Seller Storm"},
	{ID: 62, Name: "Seller Thunder"},
	{ID: 63, Name: "Seller Lightning"},
	{ID: 64, Name: "Seller Cloud"},
	{ID: 65, Name: "Seller Rain"},
	{ID: 66, Name: "Seller Wind"},
	{ID: 67, Name: "Seller Sky"},
	{ID: 68, Name: "Seller Dawn"},
	{ID: 69, Name: "Seller Dusk"},
	{ID: 70, Name: "Seller Horizon"},
	{ID: 71, Name: "Seller Terra"},
	{ID: 72, Name: "Seller Aqua"},
	{ID: 73, Name: "Seller Ignis"},
	{ID: 74, Name: "Seller Aer"},
	{ID: 75, Name: "Seller Lux"},
	{ID: 76, Name: "Seller Umbra"},
	{ID: 77, Name: "Seller Sol"},
	{ID: 78, Name: "Seller Luna"},
	{ID: 79, Name: "Seller Astra"},
	{ID: 80, Name: "Seller Argo"},
	{ID: 81, Name: "Seller Titan"},
	{ID: 82, Name: "Seller Atlas"},
	{ID: 83, Name: "Seller Orion"},
	{ID: 84, Name: "Seller Vega"},
	{ID: 85, Name: "Seller Sirius"},
	{ID: 86, Name: "Seller Polaris"},
	{ID: 87, Name: "Seller Phoenix"},
	{ID: 88, Name: "Seller Draco"},
	{ID: 89, Name: "Seller Hydra"},
	{ID: 90, Name: "Seller Pegasus"},
	{ID: 91, Name: "Seller Leo"},
	{ID: 92, Name: "Seller Aries"},
	{ID: 93, Name: "Seller Taurus"},
	{ID: 94, Name: "Seller Gemini"},
	{ID: 95, Name: "Seller Cancer"},
	{ID: 96, Name: "Seller Virgo"},
	{ID: 97, Name: "Seller Libra"},
	{ID: 98, Name: "Seller Scorpio"},
	{ID: 99, Name: "Seller Sagittarius"},
	{ID: 100, Name: "Seller Capricorn"},
}

func getAllSellers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sellers)
}

func getSellerByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["sellerId"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	for _, s := range sellers {
		if s.ID == id {
			json.NewEncoder(w).Encode(s)
			return
		}
	}

	http.Error(w, "Vendedor não encontrado", http.StatusNotFound)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/sellers", getAllSellers).Methods("GET")
	r.HandleFunc("/sellers/{sellerId}", getSellerByID).Methods("GET")

	http.ListenAndServe(":8080", r)
}
