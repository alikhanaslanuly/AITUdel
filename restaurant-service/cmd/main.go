package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcHandler "restaurant-service/internal/delivery/grpc"
	"restaurant-service/internal/repository"
	"restaurant-service/internal/usecase"
	"restaurant-service/pkg/cache"
	"restaurant-service/pkg/database"
	"restaurant-service/pkg/messaging"
	"restaurant-service/pkg/observability"
	pb "restaurant-service/proto"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()

	shutdownObs, err := observability.Init(ctx, "restaurant-service", getEnv("METRICS_PORT", "9103"))
	if err != nil {
		log.Fatalf("observability: %v", err)
	}
	defer func() {
		_ = shutdownObs(context.Background())
	}()

	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	redisClient := cache.NewRedisClient()

	natsConn := messaging.NewNATSConnection()
	defer natsConn.Close()

	repo := repository.NewPostgresRepository(db)

	restaurantUsecase := usecase.NewRestaurantUsecase(
		repo,
		redisClient,
	)

	stockUsecase := usecase.NewStockUsecase(db)

	messaging.SubscribeOrderCreated(
		natsConn,
		stockUsecase,
	)

	server := grpcHandler.NewServer(
		restaurantUsecase,
	)

	lis, err := net.Listen(
		"tcp",
		":50053",
	)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			observability.GRPCMetricsUnaryInterceptor("restaurant-service"),
		),
	)

	pb.RegisterRestaurantServiceServer(
		grpcServer,
		server,
	)

	reflection.Register(grpcServer)

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Println("shutting down restaurant-service...")
		grpcServer.GracefulStop()
	}()

	log.Println("restaurant-service gRPC running on :50053")

	if err := grpcServer.Serve(lis); err != nil {
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
