package main

import (
	"net/http"

	"cloud.redhat.com/ingress/upload"

	"github.com/go-chi/chi"
)

func LubDub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("lubdub"))
}

func main() {
	r := chi.NewRouter()
	r.Get("/", LubDub)
	r.Post("/upload", upload.Handle)
	http.ListenAndServe(":3000", r)
}
