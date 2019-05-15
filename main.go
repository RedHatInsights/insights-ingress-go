package main

import (
	"net/http"

	"cloud.redhat.com/ingress/upload"

	"github.com/RedHatInsights/platform-go-middlewares/identity"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func lubDub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("lubdub"))
}

func main() {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	)
	r.Use(
		identity.Identity,
	)
	r.Get("/", lubDub)
	r.Post("/upload", upload.NewHandler(upload.NewS3Stager("jjaggars-test")))
	r.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":3000", r)
}
