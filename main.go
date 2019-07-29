package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/redhatinsights/insights-ingress-go/announcers"
	"github.com/redhatinsights/insights-ingress-go/config"
	i "github.com/redhatinsights/insights-ingress-go/interactions/inventory"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/pipeline"
	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/stage/minio"
	"github.com/redhatinsights/insights-ingress-go/stage/s3"
	"github.com/redhatinsights/insights-ingress-go/upload"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"github.com/redhatinsights/insights-ingress-go/validators/kafka"
	"github.com/redhatinsights/insights-ingress-go/version"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/redhatinsights/platform-go-middlewares/request_id"
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
		request_id.ConfiguredRequestID("x-rh-insights-request-id"),
		middleware.RealIP,
		middleware.Recoverer,
	)

	valCh := make(chan *validators.Response)
	invCh := make(chan *validators.Response)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stager stage.Stager

	stager = &s3.Stager{
		Bucket:   cfg.StageBucket,
		Rejected: cfg.RejectBucket,
	}

	if cfg.MinioDev {
		stager = minio.GetClient(&minio.Stager{
			Bucket:   cfg.StageBucket,
			Rejected: cfg.RejectBucket,
		})
	}

	validator := kafka.New(&kafka.Config{
		Brokers:         cfg.KafkaBrokers,
		GroupID:         cfg.KafkaGroupID,
		ValidationTopic: cfg.KafkaValidationTopic,
		ValidChan:       valCh,
		InvalidChan:     invCh,
		Context:         ctx,
	}, cfg.ValidTopics...)

	inventory := &i.HTTP{
		Endpoint: cfg.InventoryURL,
	}

	tracker := announcers.NewStatusAnnouncer(&queue.ProducerConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTrackerTopic,
		Async:   true,
	})

	p := &pipeline.Pipeline{
		Stager:    stager,
		Validator: validator,
		Announcer: announcers.NewKafkaAnnouncer(&queue.ProducerConfig{
			Brokers: cfg.KafkaBrokers,
			Topic:   cfg.KafkaAvailableTopic,
		}),
		ValidChan:   valCh,
		InvalidChan: invCh,
		Tracker:     tracker,
	}

	pipelineClosed := make(chan struct{})
	go p.Start(context.Background(), pipelineClosed)

	handler := upload.NewHandler(
		stager, inventory, validator, tracker, *cfg,
	)

	var sub chi.Router = chi.NewRouter()
	if cfg.Auth {
		sub.With(identity.EnforceIdentity).Get("/", lubDub)
		sub.With(identity.EnforceIdentity, middleware.Logger).Post("/upload", handler)
	} else {
		sub.Get("/", lubDub)
		sub.With(middleware.Logger).Post("/upload", handler)
	}
	sub.With(middleware.Logger).Get("/version", version.GetVersion)

	r.Mount("/api/ingress/v1", sub)
	r.Mount("/r/insights/platform/ingress/v1", sub)
	r.Get("/", lubDub)
	r.Handle("/metrics", promhttp.Handler())

	if cfg.Profile {
		r.Mount("/debug", middleware.Profiler())
	}

	l.Log.Info("Starting service", zap.Int("port", cfg.Port))

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := srv.Shutdown(context.Background()); err != nil {
			l.Log.Fatal("HTTP Server Shutdown failed", zap.Error(err))
		}
		l.Log.Info("Canceling validation consumer")
		cancel()
		l.Log.Info("Waiting for validation pipeline to close")
		<-pipelineClosed
		l.Log.Info("Pipeline finished, canceling announcer")
		p.Announcer.Stop()
		l.Log.Info("Announcer finished, shutting down")
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		l.Log.Fatal("Service stopped", zap.Error(err))
	}

	<-idleConnsClosed
	l.Log.Info("Everything has shut down, goodbye")
}
