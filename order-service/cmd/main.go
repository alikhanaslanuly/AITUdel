package main

import (
	"log"
	"net"
	"order-service/internal/config"
	"order-service/internal/delivery/grpc"
	"order-service/internal/repository"
	"order-service/internal/usecase"
	"order-service/pkg/cache"
	"order-service/pkg/database"
	"order-service/pkg/messaging"
	pb "order-service/proto/order"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
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

	server := googlegrpc.NewServer()
	pb.RegisterOrderServiceServer(server, handler)
	reflection.Register(server)

	log.Printf("gRPC server running on :%s", cfg.GRPCPort)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
