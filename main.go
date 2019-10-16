package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"

	"github.com/redhatinsights/insights-ingress-go/announcers"
	"github.com/redhatinsights/insights-ingress-go/config"
	i "github.com/redhatinsights/insights-ingress-go/interactions/inventory"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/stage/minio"
	"github.com/redhatinsights/insights-ingress-go/stage/s3"
	"github.com/redhatinsights/insights-ingress-go/upload"
	"github.com/redhatinsights/insights-ingress-go/validators/kafka"
	"github.com/redhatinsights/insights-ingress-go/version"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/redhatinsights/platform-go-middlewares/request_id"
	"github.com/sirupsen/logrus"
)

func lubDub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("lubdub"))
}

func apiSpec(w http.ResponseWriter, r *http.Request) {
	file, err := ioutil.ReadFile("/tmp/src/openapi.yaml")
	if err != nil {
		l.Log.WithFields(logrus.Fields{"error": err}).Error("Unable to print API spec")
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(file)
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

	var stager stage.Stager

	stager = &s3.Stager{
		Bucket: cfg.StageBucket,
	}

	if cfg.MinioDev {
		stager = minio.GetClient(&minio.Stager{
			Bucket: cfg.StageBucket,
		})
	}

	validator := kafka.New(&kafka.Config{
		Brokers: cfg.KafkaBrokers,
		GroupID: cfg.KafkaGroupID,
	}, cfg.ValidTopics...)

	inventory := &i.HTTP{
		Endpoint: cfg.InventoryURL,
	}

	tracker := announcers.NewStatusAnnouncer(&queue.ProducerConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTrackerTopic,
		Async:   true,
	})

	handler := upload.NewHandler(
		stager, inventory, validator, tracker, *cfg,
	)

	var sub chi.Router = chi.NewRouter()
	if cfg.Auth {
		sub.With(identity.EnforceIdentity).Get("/", lubDub)
		sub.With(upload.ResponseMetricsMiddleware, identity.EnforceIdentity, middleware.Logger).Post("/upload", handler)
	} else {
		sub.Get("/", lubDub)
		sub.With(upload.ResponseMetricsMiddleware, middleware.Logger).Post("/upload", handler)
	}
	sub.With(middleware.Logger).Get("/version", version.GetVersion)
	sub.With(middleware.Logger).Get("/openapi.yaml", apiSpec)

	r.Mount("/api/ingress/v1", sub)
	r.Mount("/r/insights/platform/ingress/v1", sub)
	r.Get("/", lubDub)
	r.Handle("/metrics", promhttp.Handler())

	if cfg.Profile {
		r.Mount("/debug", middleware.Profiler())
	}

	l.Log.WithFields(logrus.Fields{"port": cfg.Port}).Info("Starting Service")

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
			l.Log.WithFields(logrus.Fields{"error": err}).Fatal("HTTP Server Shutdown failed")
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		l.Log.WithFields(logrus.Fields{"error": err}).Fatal("Service Stopped")
	}

	<-idleConnsClosed
	l.Log.Info("Everything has shut down, goodbye")
}
