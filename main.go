package main

import (
	"net/http"

	"cloud.redhat.com/ingress/config"
	"cloud.redhat.com/ingress/pipeline"
	"cloud.redhat.com/ingress/stage"
	"cloud.redhat.com/ingress/upload"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redhatinsights/platform-go-middlewares/identity"
)

func lubDub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("lubdub"))
}

func getPipeline() *pipeline.Pipeline {
	return &pipeline.Pipeline{
		Stager:    stage.NewS3Stager("jjaggars-test"),
		Validator: &pipeline.KafkaValidator{},
	}
}

func main() {
	cfg := config.Get()
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	)
	if cfg.Auth {
		r.Use(
			identity.Identity,
		)
	}
	r.Get("/", lubDub)
	r.Post("/upload", upload.NewHandler(getPipeline()))
	r.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":3000", r)
}
