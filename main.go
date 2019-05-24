package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redhatinsights/insights-ingress-go/announcers"
	"github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/pipeline"
	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/stage/s3"
	"github.com/redhatinsights/insights-ingress-go/upload"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"github.com/redhatinsights/insights-ingress-go/validators/kafka"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"go.uber.org/zap"
)

func lubDub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("lubdub"))
}

func main() {
	cfg := config.Get()
	l.InitLogger()
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	)
	if cfg.Auth {
		r.Use(
			identity.EnforceIdentity,
		)
	}

	valCh := make(chan *validators.Response)
	invCh := make(chan *validators.Response)

	p := &pipeline.Pipeline{
		Stager: s3.WithSession(&s3.S3Stager{
			Bucket:   cfg.StageBucket,
			Rejected: cfg.RejectBucket,
		}),
		Validator: kafka.New(&kafka.Config{
			Brokers:         cfg.KafkaBrokers,
			GroupID:         cfg.KafkaGroupID,
			ValidationTopic: cfg.KafkaValidationTopic,
			ValidChan:       valCh,
			InvalidChan:     invCh,
		}, "platform.upload.testareno"),
		Announcer: announcers.NewKafkaAnnouncer(&queue.ProducerConfig{
			Brokers: cfg.KafkaBrokers,
			Topic:   cfg.KafkaAvailableTopic,
		}),
		ValidChan:   valCh,
		InvalidChan: invCh,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go p.Start(ctx)

	r.Route("/api/ingress/v1", func(r chi.Router) {
		r.Get("/", lubDub)
		r.Post("/upload", upload.NewHandler(p))
	})
	r.Get("/", lubDub)
	r.Handle("/metrics", promhttp.Handler())
	l.Log.Info("Starting service", zap.Int("port", cfg.Port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r)
	l.Log.Fatal("Service stopped", zap.Error(err))
}
