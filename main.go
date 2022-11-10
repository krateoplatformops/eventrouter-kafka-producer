package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/krateoplatformops/eventrouter-kafka-producer/internal/handlers"
	"github.com/krateoplatformops/eventrouter-kafka-producer/internal/handlers/eventrouter"
	"github.com/krateoplatformops/eventrouter-kafka-producer/internal/middlewares"
	"github.com/krateoplatformops/eventrouter-kafka-producer/internal/support"
	"github.com/rs/zerolog"

	"github.com/gorilla/mux"

	"github.com/Shopify/sarama"
	kafka_sarama "github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
)

const (
	serviceName = "KafkaEventRouterConsumer"
)

var (
	Version string
	Build   string
)

func main() {
	// Flags
	debug := flag.Bool("debug", support.EnvBool("EVENTROUTER_KAFKA_PRODUCER_DEBUG", true), "dump verbose output")
	servicePort := flag.Int("port", support.EnvInt("EVENTROUTER_KAFKA_PRODUCER_PORT", 8080), "port to listen on")
	brokers := flag.String("brokers",
		support.EnvString("EVENTROUTER_KAFKA_PRODUCER_BROKERS", "127.0.0.1:9092"), "Kafka brokers comma separated list")
	topic := flag.String("topic",
		support.EnvString("EVENTROUTER_KAFKA_PRODUCER_TOPIC", "test-topic"), "send events to this Kafka topic")

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Initialize the logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Default level for this log is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log := zerolog.New(os.Stdout).With().
		Str("service", serviceName).
		Timestamp().
		Logger()

	if log.Debug().Enabled() {
		log.Debug().
			Str("version", Version).
			Str("build", Build).
			Str("debug", fmt.Sprintf("%t", *debug)).
			Str("port", fmt.Sprintf("%d", *servicePort)).
			Str("brokers", *brokers).
			Msgf("Configuration values for %s.", serviceName)
	}

	// Server Mux
	mux := mux.NewRouter()

	// HealtZ endpoint
	healthy := int32(0)
	mux.Handle("/healthz", middlewares.CorrelationID(
		handlers.HealtHandler(&healthy, serviceName, Version),
	))

	// Setup Kafka sender
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0

	sender, err := kafka_sarama.NewSender(strings.Split(*brokers, ","), saramaConfig, *topic)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka sender")
	}
	defer sender.Close(context.Background())

	mux.Handle("/handle", middlewares.Logger(log)(
		middlewares.CorrelationID(
			eventrouter.Handler(sender, *debug),
		),
	)).Methods(http.MethodPost)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *servicePort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 40 * time.Second,
		IdleTimeout:  20 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), []os.Signal{
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}...)
	defer stop()

	go func() {
		atomic.StoreInt32(&healthy, 1)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("could not listen on %s", server.Addr)
		}
	}()

	// Listen for the interrupt signal.
	log.Info().Msgf("server is ready to handle requests at @ %s", server.Addr)
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info().Msg("server is shutting down gracefully, press Ctrl+C again to force")
	atomic.StoreInt32(&healthy, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server gracefully stopped")
}
