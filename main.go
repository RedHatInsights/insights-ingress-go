package main

import (
	"context"
	"log"
	"net/http"

	"cloud.redhat.com/ingress/announcers"
	"cloud.redhat.com/ingress/config"
	"cloud.redhat.com/ingress/pipeline"
	"cloud.redhat.com/ingress/stage/s3"
	"cloud.redhat.com/ingress/upload"
	"cloud.redhat.com/ingress/validators"

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

func main() {
	cfg := config.Get()
	log.Printf("cfg: %v", cfg)
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

	valCh := make(chan *validators.Response)
	invCh := make(chan *validators.Response)

	p := &pipeline.Pipeline{
		Stager: s3.New("jjaggars-test"),
		Validator: validators.NewKafkaValidator(&validators.KafkaConfig{
			Brokers:         cfg.KafkaBrokers,
			GroupID:         cfg.KafkaGroupID,
			AvailableTopic:  cfg.KafkaAvailableTopic,
			ValidationTopic: cfg.KafkaValidationTopic,
			ValidChan:       valCh,
			InvalidChan:     invCh,
		}, "platform.upload.testareno"),
		Announcer:   &announcers.Fake{},
		ValidChan:   valCh,
		InvalidChan: invCh,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go p.Start(ctx)

	r.Get("/", lubDub)
	r.Post("/upload", upload.NewHandler(p))
	r.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":3000", r)
}
