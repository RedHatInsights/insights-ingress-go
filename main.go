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
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/stage/minio"
	"github.com/redhatinsights/insights-ingress-go/stage/s3"
	"github.com/redhatinsights/insights-ingress-go/track"
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
	file, err := ioutil.ReadFile("/var/tmp/openapi.json")
	if err != nil {
		l.Log.WithFields(logrus.Fields{"error": err}).Error("Unable to print API spec")
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(file)
}

func main() {
	l.InitLogger()
	cfg := config.Get()
	r := chi.NewRouter()
	mr := chi.NewRouter()
	r.Use(
		request_id.ConfiguredRequestID("x-rh-insights-request-id"),
		middleware.RealIP,
		middleware.Recoverer,
	)

	var stager stage.Stager
	if cfg.MinioEndpoint != ""{
		stager = minio.GetClient(&minio.Stager{
			Bucket: cfg.StageBucket,
		})
	} else {
		stager = &s3.Stager{
			Bucket: cfg.StageBucket,
		}
	}

	validator := kafka.New(&kafka.Config{
		Brokers: cfg.KafkaBrokers,
		GroupID: cfg.KafkaGroupID,
	}, cfg.ValidTopics...)

	tracker := announcers.NewStatusAnnouncer(&queue.ProducerConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTrackerTopic,
		Async:   true,
	})

	handler := upload.NewHandler(
		stager, validator, tracker, *cfg,
	)

	trackEndpoint := track.NewHandler(
		*cfg,
	)

	var sub chi.Router = chi.NewRouter()
	if cfg.Auth {
		sub.With(identity.EnforceIdentity).Get("/", lubDub)
		sub.With(upload.ResponseMetricsMiddleware, identity.EnforceIdentity, middleware.Logger).Post("/upload", handler)
		sub.With(identity.EnforceIdentity).Get("/track/{requestID}", trackEndpoint)
	} else {
		sub.Get("/", lubDub)
		sub.With(upload.ResponseMetricsMiddleware, middleware.Logger).Post("/upload", handler)
	}
	sub.With(middleware.Logger).Get("/version", version.GetVersion)
	sub.With(middleware.Logger).Get("/openapi.json", apiSpec)

	r.Mount("/api/ingress/v1", sub)
	r.Mount("/r/insights/platform/ingress/v1", sub)
	r.Get("/", lubDub)
	mr.Get("/", lubDub)
	mr.Handle("/metrics", promhttp.Handler())

	if cfg.Profile {
		r.Mount("/debug", middleware.Profiler())
	}

	l.Log.WithFields(logrus.Fields{"Web Port": cfg.WebPort}).Info("Starting Service")

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.WebPort),
		Handler: r,
	}

	l.Log.WithFields(logrus.Fields{"Metrics Port": cfg.MetricsPort}).Info("Starting Service")

	msrv := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler: mr,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := srv.Shutdown(context.Background()); err != nil {
			l.Log.WithFields(logrus.Fields{"error": err}).Fatal("HTTP Server Shutdown failed")
		}
		if err := msrv.Shutdown(context.Background()); err != nil {
			l.Log.WithFields(logrus.Fields{"error": err}).Fatal("HTTP Server Shutdown failed")
		}
		close(idleConnsClosed)
	}()

	// create and expose the version information as a prometheus metric
	version.ExposeVersion()

	go func() {

		if err := msrv.ListenAndServe(); err != http.ErrServerClosed {
			l.Log.WithFields(logrus.Fields{"error": err}).Fatal("Metrics Service Stopped")
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		l.Log.WithFields(logrus.Fields{"error": err}).Fatal("Service Stopped")
	}

	<-idleConnsClosed
	l.Log.Info("Everything has shut down, goodbye")
}
