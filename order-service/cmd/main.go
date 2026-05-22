package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"order-service/internal/config"
	"order-service/internal/delivery/grpc"
	"order-service/internal/repository"
	"order-service/internal/usecase"
	"order-service/pkg/cache"
	"order-service/pkg/database"
	"order-service/pkg/messaging"
	"order-service/pkg/observability"
	pb "order-service/proto/order"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()

	shutdownObs, err := observability.Init(ctx, "order-service", getEnv("METRICS_PORT", "9101"))
	if err != nil {
		log.Fatalf("observability: %v", err)
	}
	defer func() {
		_ = shutdownObs(context.Background())
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.NewPostgres(cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations done")

	redisClient, err := cache.NewRedis(cfg.RedisAddr())
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	natsClient, err := messaging.NewNats(cfg.NatsURL)
	if err != nil {
		log.Fatalf("nats: %v", err)
	}
	defer natsClient.Close()

	orderRepo := repository.NewOrderRepository(db)
	promoRepo := repository.NewPromoRepository(db)

	orderUC := usecase.NewOrderUsecase(orderRepo, promoRepo, redisClient, natsClient)
	cartUC := usecase.NewCartUsecase(redisClient)

	handler := grpc.NewOrderHandler(orderUC, cartUC)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := googlegrpc.NewServer(
		googlegrpc.StatsHandler(otelgrpc.NewServerHandler()),
		googlegrpc.ChainUnaryInterceptor(
			observability.GRPCMetricsUnaryInterceptor("order-service"),
		),
	)

	pb.RegisterOrderServiceServer(server, handler)
	reflection.Register(server)

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Println("shutting down order-service...")
		server.GracefulStop()
	}()

	log.Printf("order-service gRPC running on :%s", cfg.GRPCPort)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
