package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"

	"github.com/redhatinsights/insights-ingress-go/internal/announcers"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/insights-ingress-go/internal/featureflags"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/redhatinsights/insights-ingress-go/internal/queue"
	"github.com/redhatinsights/insights-ingress-go/internal/stage/s3compat"
	"github.com/redhatinsights/insights-ingress-go/internal/track"
	"github.com/redhatinsights/insights-ingress-go/internal/upload"
	"github.com/redhatinsights/insights-ingress-go/internal/validators/kafka"
	"github.com/redhatinsights/insights-ingress-go/internal/version"

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
	cfg := config.Get()
	l.InitLogger(cfg)
	r := chi.NewRouter()
	mr := chi.NewRouter()
	r.Use(
		request_id.ConfiguredRequestID("x-rh-insights-request-id"),
		middleware.RealIP,
		middleware.Recoverer,
	)

	stager := s3compat.GetClient(cfg, &s3compat.Stager{
		Bucket: cfg.StorageConfig.StageBucket,
	})

	kafkaCfg := kafka.Config{
		Brokers: cfg.KafkaConfig.KafkaBrokers,
		GroupID: cfg.KafkaConfig.KafkaGroupID,
	}

	producerCfg := queue.ProducerConfig{
		Brokers:              cfg.KafkaConfig.KafkaBrokers,
		Topic:                cfg.KafkaConfig.KafkaTrackerTopic,
		Async:                true,
		KafkaDeliveryReports: cfg.KafkaConfig.KafkaDeliveryReports,
	}

	// Kafka SSL Config
	if cfg.KafkaConfig.KafkaSSLConfig != (config.KafkaSSLCfg{}) {
		kafkaCfg.CA = cfg.KafkaConfig.KafkaSSLConfig.KafkaCA
		kafkaCfg.Username = cfg.KafkaConfig.KafkaSSLConfig.KafkaUsername
		kafkaCfg.Password = cfg.KafkaConfig.KafkaSSLConfig.KafkaPassword
		kafkaCfg.SASLMechanism = cfg.KafkaConfig.KafkaSSLConfig.SASLMechanism
		kafkaCfg.Protocol = cfg.KafkaConfig.KafkaSSLConfig.Protocol
		producerCfg.CA = cfg.KafkaConfig.KafkaSSLConfig.KafkaCA
		producerCfg.Username = cfg.KafkaConfig.KafkaSSLConfig.KafkaUsername
		producerCfg.Password = cfg.KafkaConfig.KafkaSSLConfig.KafkaPassword
		producerCfg.SASLMechanism = cfg.KafkaConfig.KafkaSSLConfig.SASLMechanism
		producerCfg.Protocol = cfg.KafkaConfig.KafkaSSLConfig.Protocol
	}

	validator := kafka.New(&kafkaCfg, cfg.KafkaConfig.ValidTopics...)

	tracker := announcers.NewStatusAnnouncer(&producerCfg)

	var featureFlags featureflags.FeatureFlagClient
	var err error

	if cfg.FeatureFlagsConfig != (config.FeatureFlagCfg{}) {
		featureFlags, err = featureflags.NewFeatureFlagClient("unleash", cfg)
		if err != nil {
			l.Log.Error(err)
		}
	} else {
		featureFlags, err = featureflags.NewFeatureFlagClient("fake", cfg)
	}
	featureFlags.InitializeClient()

	handler := upload.NewHandler(
		stager, validator, tracker, featureFlags, *cfg,
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
