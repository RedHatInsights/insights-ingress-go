package main

import (
	"net/http"

	"cloud.redhat.com/ingress/upload"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func LubDub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("lubdub"))
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Get("/", LubDub)
	r.Post("/upload", upload.Handle)
	http.ListenAndServe(":3000", r)
}
