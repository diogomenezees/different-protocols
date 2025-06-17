package main

import (
	"net/http"
	"strconv"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/gorilla/mux"
)

type Image struct {
	ID  int    `msgpack:"id"`
	URL string `msgpack:"url"`
}

var images = []Image{}

func init() {
	for i := 1; i <= 100; i++ {
		images = append(images, Image{ID: i, URL: "https://example.com/image" + strconv.Itoa(i) + ".jpg"})
	}
}

func getAllImages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-msgpack")
	msgpack.NewEncoder(w).Encode(images)
}

func getImageByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-msgpack")

	vars := mux.Vars(r)
	idStr := vars["imageId"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	for _, img := range images {
		if img.ID == id {
			msgpack.NewEncoder(w).Encode(img)
			return
		}
	}

	http.Error(w, "Imagem não encontrada", http.StatusNotFound)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/images", getAllImages).Methods("GET")
	r.HandleFunc("/images/{imageId}", getImageByID).Methods("GET")

	http.ListenAndServe(":8080", r)
}
