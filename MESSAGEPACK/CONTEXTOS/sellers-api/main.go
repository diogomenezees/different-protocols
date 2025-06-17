package main

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vmihailenco/msgpack/v5"
)

type Seller struct {
	ID   int    `msgpack:"id"`
	Name string `msgpack:"name"`
}

var sellers = []Seller{}

func init() {
	for i := 1; i <= 100; i++ {
		sellers = append(sellers, Seller{
			ID:   i,
			Name: "Seller " + strconv.Itoa(i),
		})
	}
}

func getAllSellers(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	_ = enc.Encode(sellers)
	w.Header().Set("Content-Type", "application/x-msgpack")
	w.Write(buf.Bytes())
}

func getSellerByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["sellerId"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	for _, s := range sellers {
		if s.ID == id {
			var buf bytes.Buffer
			enc := msgpack.NewEncoder(&buf)
			_ = enc.Encode(s)
			w.Header().Set("Content-Type", "application/x-msgpack")
			w.Write(buf.Bytes())
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
