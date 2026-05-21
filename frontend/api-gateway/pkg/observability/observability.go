package observability

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aitudel_requests_total",
			Help: "Total number of requests by service, protocol, method and status.",
		},
		[]string{"service", "protocol", "method", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aitudel_request_duration_seconds",
			Help:    "Request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "protocol", "method", "status"},
	)
)

func init() {
	prometheus.MustRegister(requestTotal)
	prometheus.MustRegister(requestDuration)
}

func Init(ctx context.Context, serviceName string, metricsPort string) (func(context.Context) error, error) {
	if err := os.MkdirAll("../../logs", 0755); err != nil {
		return nil, err
	}

	logFilePath := filepath.Join("../../logs", serviceName+".log")

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	writer := io.MultiWriter(os.Stdout, file)

	logger := slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	log.SetOutput(writer)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4319")),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	go startMetricsServer(metricsPort)

	slog.Info("observability initialized",
		"service", serviceName,
		"metrics_port", metricsPort,
		"otel_endpoint", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4319"),
	)

	return func(ctx context.Context) error {
		err := tracerProvider.Shutdown(ctx)
		_ = file.Close()
		return err
	}, nil
}

func startMetricsServer(port string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("metrics server started", "port", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("metrics server failed", "error", err)
	}
}

func GRPCMetricsUnaryInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		code := status.Code(err).String()
		method := info.FullMethod

		requestTotal.WithLabelValues(serviceName, "grpc", method, code).Inc()
		requestDuration.WithLabelValues(serviceName, "grpc", method, code).Observe(time.Since(start).Seconds())

		if err != nil {
			slog.Error("grpc request failed",
				"service", serviceName,
				"method", method,
				"status", code,
				"duration_ms", time.Since(start).Milliseconds(),
				"error", err.Error(),
			)
		} else {
			slog.Info("grpc request completed",
				"service", serviceName,
				"method", method,
				"status", code,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}

		return resp, err
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func HTTPMetricsMiddleware(serviceName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		status := strconv.Itoa(recorder.statusCode)
		method := r.Method + " " + r.URL.Path

		requestTotal.WithLabelValues(serviceName, "http", method, status).Inc()
		requestDuration.WithLabelValues(serviceName, "http", method, status).Observe(time.Since(start).Seconds())

		slog.Info("http request completed",
			"service", serviceName,
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func TraceHTTPHandler(operationName string, next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, operationName)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
