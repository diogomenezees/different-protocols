package main

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vmihailenco/msgpack/v5"
)

type Brand struct {
	ID          int    `msgpack:"id"`
	Name        string `msgpack:"name"`
	Description string `msgpack:"description"`
	Country     string `msgpack:"country"`
	Active      bool   `msgpack:"active"`
}

var brands = []Brand{}

func init() {
	for i := 1; i <= 100; i++ {
		brands = append(brands, Brand{
			ID:          i,
			Name:        "Brand " + strconv.Itoa(i),
			Description: "Descrição da marca " + strconv.Itoa(i),
			Country:     "País " + strconv.Itoa(i%5+1),
			Active:      i%2 == 0,
		})
	}
}

func getAllBrands(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	_ = enc.Encode(brands)
	w.Header().Set("Content-Type", "application/x-msgpack")
	w.Write(buf.Bytes())
}

func getBrandByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["brandId"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	for _, b := range brands {
		if b.ID == id {
			var buf bytes.Buffer
			enc := msgpack.NewEncoder(&buf)
			_ = enc.Encode(b)
			w.Header().Set("Content-Type", "application/x-msgpack")
			w.Write(buf.Bytes())
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
